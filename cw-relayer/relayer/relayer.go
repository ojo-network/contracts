package relayer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	psync "github.com/ojo-network/cw-relayer/pkg/sync"
	"github.com/ojo-network/cw-relayer/relayer/client"
)

var (
	// RateFactor is used to convert ojo prices to contract-compatible values.
	RateFactor   = types.NewDec(10).Power(9)
	noRates      = fmt.Errorf("no rates found")
	noMedians    = fmt.Errorf("median deviations empty")
	noDeviations = fmt.Errorf("deviation deviations empty")
)

// Relayer defines a structure that queries prices from ojo and publishes prices to wasm contract.
type Relayer struct {
	logger    zerolog.Logger
	closer    *psync.Closer
	queryRPCS []string

	relayerAddress  string
	contractAddress string

	tickDuration time.Duration

	resolveDuration time.Duration
	queryTimeout    time.Duration

	// if missedCounter >= missedThreshold, force relay prices (bypasses timing restrictions)
	missedCounter     int64
	missedThreshold   int64
	timeoutHeight     int64
	medianDuration    int64
	deviationDuration int64

	ignoreMedianErrors bool

	oracleEvent   chan struct{}
	contractEvent chan struct{}

	rwmutex sync.RWMutex

	cs *client.ContractSubscribe
	ps *PriceService

	msg chan types.Msg
}

// New returns an instance of the relayer.
func New(
	logger zerolog.Logger,
	cs *client.ContractSubscribe,
	ps *PriceService,
	relayerAddress string,
	contractAddress string,
	timeoutHeight int64,
	tickDuration time.Duration,
	oracleEvent chan struct{},
	msg chan types.Msg,
) *Relayer {
	return &Relayer{
		logger:          logger.With().Str("module", "relayer").Logger(),
		relayerAddress:  relayerAddress,
		contractAddress: contractAddress,
		timeoutHeight:   timeoutHeight,

		closer:       psync.NewCloser(),
		tickDuration: tickDuration,
		oracleEvent:  oracleEvent,
		cs:           cs,
		ps:           ps,
		msg:          msg,
	}
}

func (r *Relayer) Start(ctx context.Context) error {

	// auto restart
	ticker := time.NewTicker(r.tickDuration)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			r.closer.Close()

		case <-ticker.C:
			r.logger.Info().Msg("contract events")
			getRequests := r.cs.GetPriceRequest()
			if len(getRequests) > 0 {
				err := r.processRequests(ctx, getRequests)
				if err != nil {
					r.logger.Err(err).Send()
				}
			}
		}
	}
}

func (r *Relayer) processRequests(ctx context.Context, requests map[string][]client.PriceRequest) error {
	denomList := make([]string, len(requests))
	i := 0
	for denom, _ := range requests {
		denomList[i] = denom
		i++
	}

	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		prices := r.ps.GetPrices(denomList)
		for denom, reqs := range requests {
			for _, req := range reqs {
				if req.Event == client.RequestRate {
					price, found := prices[denom]
					if !found {
						r.logger.Debug().Str("denom", denom).Msg("Denom price not found")
						continue
					}

					callback := CallbackRate{
						CallbackData{
							RequestID:    req.RequestID,
							Symbol:       denom,
							SymbolRate:   price.Price,
							LastUpdated:  price.Timestamp,
							CallbackData: []byte(req.CallbackData),
						},
					}

					tx := Execute{
						Callback: callback,
					}

					msg, err := genMsg(r.relayerAddress, r.contractAddress, tx)
					if err != nil {
						return err
					}

					r.msg <- msg
				}
			}
		}
		return nil
	})

	// Go routine for handling medians
	g.Go(func() error {
		medians := r.ps.GetMedians(denomList)
		for denom, reqs := range requests {
			for _, req := range reqs {
				if req.Event == client.RequestMedian {
					median, found := medians[denom]
					if !found {
						r.logger.Debug().Str("denom", denom).Msg("Denom Medians not found")
						continue
					}

					callback := CallbackMedian{
						Req: CallbackDataMedian{
							RequestID:    req.RequestID,
							Symbol:       denom,
							SymbolRates:  median.Median,
							LastUpdated:  median.Timestamp,
							CallbackData: nil,
						},
					}

					tx := Execute{
						Callback: callback,
					}

					msg, err := genMsg(r.relayerAddress, r.contractAddress, tx)
					if err != nil {
						return err
					}

					r.msg <- msg
				}
			}
		}
		return nil
	})

	// Go routine for handling deviations
	g.Go(func() error {
		deviations := r.ps.GetDeviations(denomList)
		for denom, reqs := range requests {
			for _, req := range reqs {
				if req.Event == client.RequestDeviation {
					deviation, found := deviations[denom]
					if !found {
						r.logger.Debug().Str("denom", denom).Msg("Denom Deviation not found")
						continue
					}

					callback := CallbackRate{
						Req: CallbackData{
							RequestID:    req.RequestID,
							Symbol:       denom,
							SymbolRate:   deviation.Deviation,
							LastUpdated:  deviation.Timestamp,
							CallbackData: []byte(req.CallbackData),
						},
					}

					tx := Execute{
						Callback: callback,
					}

					msg, err := genMsg(r.relayerAddress, r.contractAddress, tx)
					if err != nil {
						return err
					}

					r.msg <- msg
				}
			}
		}
		return nil
	})

	return g.Wait()
}

// Stop stops the relayer process and waits for it to gracefully exit.
func (r *Relayer) Stop() {
	r.closer.Close()
	<-r.closer.Done()
}

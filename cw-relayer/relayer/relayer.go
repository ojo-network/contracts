package relayer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	psync "github.com/ojo-network/cw-relayer/pkg/sync"
	"github.com/ojo-network/cw-relayer/relayer/client"
	"github.com/rs/zerolog"
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
	logger zerolog.Logger
	closer *psync.Closer

	relayerClient   client.RelayerClient
	queryRPCS       []string
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
}

// New returns an instance of the relayer.
func New(
	logger zerolog.Logger,
	oc client.RelayerClient,
	cs *client.ContractSubscribe,
	ps *PriceService,
	contractAddress string,
	timeoutHeight int64,
	tickDuration time.Duration,
	oracleEvent chan struct{},
) *Relayer {
	return &Relayer{
		logger:          logger.With().Str("module", "relayer").Logger(),
		relayerClient:   oc,
		contractAddress: contractAddress,
		timeoutHeight:   timeoutHeight,

		closer:       psync.NewCloser(),
		tickDuration: tickDuration,
		oracleEvent:  oracleEvent,
		cs:           cs,
		ps:           ps,
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
				msgs, err := r.processRequests(getRequests)
				if err != nil {
					r.logger.Err(err).Send()
				}

				err = r.tick(msgs)
				if err != nil {
					r.logger.Err(err).Send()
				}
			}
		}
	}
}

func (r *Relayer) processRequests(requests map[string][]client.PriceRequest) ([]types.Msg, error) {
	var msgs []types.Msg
	denomList := make([]string, len(requests))
	i := 0
	for denom, _ := range requests {
		denomList[i] = denom
		i++
	}

	prices := r.ps.GetPrices(denomList)
	medians := r.ps.GetMedians(denomList)
	deviations := r.ps.GetDeviations(denomList)
	for denom, reqs := range requests {
		for _, req := range reqs {
			var callback interface{}
			switch req.Event {
			case client.RequestRate:
				price, found := prices[denom]
				if !found {
					// skipping request symbol not found
					r.logger.Debug().Str("denom", denom).Msg("Denom price not found")
					continue
				}

				callback = CallbackRate{
					CallbackData{
						RequestID:    req.RequestID,
						Symbol:       denom,
						SymbolRate:   price.Price,
						LastUpdated:  price.Timestamp,
						CallbackData: []byte(req.CallbackData),
					},
				}
			case client.RequestMedian:
				median, found := medians[denom]
				if !found {
					// skipping request symbol not found
					r.logger.Debug().Str("denom", denom).Msg("Denom Medians not found")
					continue
				}

				callback = CallbackMedian{Req: CallbackDataMedian{
					RequestID:    req.RequestID,
					Symbol:       denom,
					SymbolRates:  median.Median,
					LastUpdated:  median.Timestamp,
					CallbackData: nil,
				}}

			case client.RequestDeviation:
				deviation, found := deviations[denom]
				if !found {
					// skipping request symbol not found
					r.logger.Debug().Str("denom", denom).Msg("Denom Deviation not found")
					continue
				}

				callback = CallbackDeviation{Req: CallbackData{
					RequestID:    req.RequestID,
					Symbol:       denom,
					SymbolRate:   deviation.Deviation,
					LastUpdated:  deviation.Timestamp,
					CallbackData: []byte(req.CallbackData),
				}}
			}

			//TODO filter resolve times
			tx := Execute{
				Callback: callback,
			}

			msg, err := genMsg(r.relayerClient.RelayerAddrString, r.contractAddress, tx)
			if err != nil {
				return nil, err
			}

			msgs = append(msgs, msg)
		}
	}

	return msgs, nil
}

// Stop stops the relayer process and waits for it to gracefully exit.
func (r *Relayer) Stop() {
	r.closer.Close()
	<-r.closer.Done()
}

func (r *Relayer) tick(msgs []types.Msg) error {
	return r.relayerClient.BroadcastTx(r.timeoutHeight, msgs...)
}

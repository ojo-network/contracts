package relayer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	oracletypes "github.com/ojo-network/ojo/x/oracle/types"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	psync "github.com/ojo-network/cw-relayer/pkg/sync"
	"github.com/ojo-network/cw-relayer/relayer/client"
	"github.com/ojo-network/cw-relayer/tools"
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

	relayerClient      client.RelayerClient
	queryRPCS          []string
	contractAddress    string
	requestID          uint64
	medianRequestID    uint64
	deviationRequestID uint64

	exchangeRates        map[string]types.DecCoin
	historicalMedians    types.DecCoins
	historicalDeviations types.DecCoins
	resolveDuration      time.Duration
	queryTimeout         time.Duration

	// if missedCounter >= missedThreshold, force relay prices (bypasses timing restrictions)
	missedCounter     int64
	missedThreshold   int64
	timeoutHeight     int64
	medianDuration    int64
	deviationDuration int64
	maxQueryRetries   int64
	queryRetries      int64
	index             int

	ignoreMedianErrors bool

	oracleEvent   chan struct{}
	contractEvent chan struct{}

	rwmutex sync.RWMutex

	cs *client.ContractSubscribe
}

type AutoRestartConfig struct {
	AutoRestart bool
	Denom       string
	SkipError   bool
}

// New returns an instance of the relayer.
func New(
	logger zerolog.Logger,
	oc client.RelayerClient,
	cs *client.ContractSubscribe,
	contractAddress string,
	timeoutHeight int64,
	maxQueryRetries int64,
	queryTimeout time.Duration,
	oracleEvent chan struct{},
	queryRPCS []string,
) *Relayer {
	return &Relayer{
		queryRPCS:       queryRPCS,
		logger:          logger.With().Str("module", "relayer").Logger(),
		relayerClient:   oc,
		contractAddress: contractAddress,
		timeoutHeight:   timeoutHeight,
		queryTimeout:    queryTimeout,
		maxQueryRetries: maxQueryRetries,
		exchangeRates:   make(map[string]types.DecCoin),
		closer:          psync.NewCloser(),
		oracleEvent:     oracleEvent,
		cs:              cs,
	}
}

func (r *Relayer) Start(ctx context.Context, tickDuration time.Duration) error {

	// auto restart
	ticker := time.NewTicker(tickDuration)
	for {
		select {
		case <-ctx.Done():
			r.closer.Close()

		case <-r.oracleEvent:
			r.logger.Info().Msg("oracle price tick")
			err := r.setDenomPrices(ctx)
			if err != nil {
				r.logger.Error().Err(err).Msg("error configuring prices")
			}

		case <-ticker.C:
			r.logger.Info().Msg("contract events")
			getRequests := r.cs.GetPriceRequest()
			if len(getRequests) > 0 {
				if len(r.exchangeRates) == 0 {
					err := r.setDenomPrices(ctx)
					if err != nil {
						return err
					}
				}
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
	for symbol, reqs := range requests {
		for _, req := range reqs {
			price, found := r.exchangeRates[symbol]
			if !found {
				// skipping request symbol not found
				continue
			}
			//TODO filter resolve times
			tx := Execute{
				Callback: Callback{
					CallbackData{
						RequestID:    req.RequestID,
						Symbol:       symbol,
						SymbolRate:   price.Amount.Mul(RateFactor).TruncateInt().String(),
						ResolveTime:  "0",
						CallbackData: []byte(req.CallbackData),
					},
				},
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

// incrementIndex increases index to switch to different query rpc
func (r *Relayer) increment() {
	r.queryRetries += 1
	r.index = (r.index + 1) % len(r.queryRPCS)
	r.logger.Info().Int("rpc index", r.index).Msg("switching query rpc")
}

func (r *Relayer) setDenomPrices(ctx context.Context) error {
	if r.queryRetries > r.maxQueryRetries {
		r.queryRetries = 0
		return fmt.Errorf("retry threshold exceeded")
	}

	grpcConn, err := grpc.Dial(
		r.queryRPCS[r.index],
		// the Cosmos SDK doesn't support any transport security mechanism
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(tools.DialerFunc),
	)

	// retry or switch rpc
	if err != nil {
		r.increment()
		return r.setDenomPrices(ctx)
	}

	defer grpcConn.Close()

	queryClient := oracletypes.NewQueryClient(grpcConn)

	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	queryResponse, err := queryClient.ExchangeRates(ctx, &oracletypes.QueryExchangeRates{})
	// assuming an issue with rpc if exchange rates are empty
	if err != nil || queryResponse.ExchangeRates.Empty() {
		r.logger.Debug().Msg("error querying exchange rates")
		r.increment()
		return r.setDenomPrices(ctx)
	}

	for _, coin := range queryResponse.ExchangeRates {
		r.exchangeRates[coin.Denom] = coin
	}

	var mu sync.Mutex
	g, _ := errgroup.WithContext(ctx)
	g.Go(
		func() error {
			deviationsQueryResponse, err := queryClient.MedianDeviations(ctx, &oracletypes.QueryMedianDeviations{})
			if err != nil {
				return err
			}

			if len(deviationsQueryResponse.MedianDeviations) == 0 {
				return noDeviations
			}

			deviations := make([]types.DecCoin, len(deviationsQueryResponse.MedianDeviations))
			for i, priceStamp := range deviationsQueryResponse.MedianDeviations {
				deviations[i] = *priceStamp.ExchangeRate
			}

			mu.Lock()
			r.historicalDeviations = deviations
			mu.Unlock()

			return nil
		},
	)

	g.Go(
		func() error {
			medianQueryResponse, err := queryClient.Medians(ctx, &oracletypes.QueryMedians{})
			if err != nil {
				return err
			}

			if len(medianQueryResponse.Medians) == 0 {
				return noMedians
			}

			medians := make([]types.DecCoin, len(medianQueryResponse.Medians))
			for i, priceStamp := range medianQueryResponse.Medians {
				medians[i] = *priceStamp.ExchangeRate
			}

			mu.Lock()
			r.historicalMedians = medians
			mu.Unlock()

			return nil
		},
	)

	return g.Wait()
}

func (r *Relayer) tick(msgs []types.Msg) error {
	return r.relayerClient.BroadcastTx(r.timeoutHeight, msgs...)
}

package relayer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
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
	RateFactor = types.NewDec(10).Power(9)
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

	exchangeRates        types.DecCoins
	historicalMedians    types.DecCoins
	historicalDeviations types.DecCoins
	resolveDuration      time.Duration
	queryTimeout         time.Duration

	// if missedCounter >= missedThreshold, force relay prices (bypasses timing restrictions)
	missedCounter   int64
	missedThreshold int64
	timeoutHeight   int64
	medianDuration  int64
	maxQueryRetries int64
	queryRetries    int64
	index           int

	event  chan struct{}
	config AutoRestartConfig
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
	contractAddress string,
	timeoutHeight int64,
	missedThreshold int64,
	maxQueryRetries int64,
	medianDuration int64,
	resolveDuration time.Duration,
	queryTimeout time.Duration,
	requestID uint64,
	medianRequestID uint64,
	deviationRequestID uint64,
	config AutoRestartConfig,
	event chan struct{},
	queryRPCS []string,
) *Relayer {
	return &Relayer{
		queryRPCS:          queryRPCS,
		logger:             logger.With().Str("module", "relayer").Logger(),
		relayerClient:      oc,
		contractAddress:    contractAddress,
		missedThreshold:    missedThreshold,
		timeoutHeight:      timeoutHeight,
		queryTimeout:       queryTimeout,
		medianDuration:     medianDuration,
		resolveDuration:    resolveDuration,
		requestID:          requestID,
		medianRequestID:    medianRequestID,
		deviationRequestID: deviationRequestID,
		maxQueryRetries:    maxQueryRetries,
		closer:             psync.NewCloser(),
		event:              event,
		config:             config,
	}
}

func (r *Relayer) Start(ctx context.Context) error {
	// auto restart
	if r.config.AutoRestart {
		err := r.restart(ctx)
		if err != nil {
			r.logger.Error().Err(err).Msg("error auto restarting relayer")

			// return error if skip error is false
			if !r.config.SkipError {
				return err
			}
		}

		r.logger.Info().
			Uint64("request id", r.requestID).
			Uint64("median request id", r.medianRequestID).
			Uint64("deviation request id", r.deviationRequestID).Msg("relayer state startup successful")
	}

	for {
		select {
		case <-ctx.Done():
			r.closer.Close()

		case <-r.event:
			r.logger.Debug().Msg("relayer tick")
			startTime := time.Now()
			if err := r.tick(ctx); err != nil {
				telemetry.IncrCounter(1, "failure", "tick")
				r.logger.Err(err).Msg("relayer tick failed")
			}

			telemetry.MeasureSince(startTime, "runtime", "tick")
			telemetry.IncrCounter(1, "new", "tick")
		}
	}
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

// restart queries wasmd chain to fetch latest request, median request and deviation request id
func (r *Relayer) restart(ctx context.Context) error {
	response, err := r.relayerClient.BroadcastContractQuery(ctx, r.config.Denom)
	if err != nil {
		return err
	}

	r.requestID = response.PriceID
	r.deviationRequestID = response.DeviationID
	r.medianRequestID = response.MedianID

	return nil
}

func (r *Relayer) setDenomPrices(ctx context.Context, postMedian bool) error {
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
		return r.setDenomPrices(ctx, postMedian)
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
		return r.setDenomPrices(ctx, postMedian)
	}

	r.exchangeRates = queryResponse.ExchangeRates

	var mu sync.Mutex
	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		deviationsQueryResponse, err := queryClient.MedianDeviations(ctx, &oracletypes.QueryMedianDeviations{})
		if err != nil {
			return err
		}

		if deviationsQueryResponse.MedianDeviations.Empty() {
			return fmt.Errorf("median deviations empty")
		}

		mu.Lock()
		r.historicalDeviations = deviationsQueryResponse.MedianDeviations
		mu.Unlock()

		return nil
	})

	if postMedian {
		g.Go(func() error {
			medianQueryResponse, err := queryClient.Medians(ctx, &oracletypes.QueryMedians{})
			if err != nil {
				return err
			}

			if medianQueryResponse.Medians.Empty() {
				return fmt.Errorf("median rates empty")
			}

			mu.Lock()
			r.historicalMedians = medianQueryResponse.Medians
			mu.Unlock()

			return nil
		})
	}

	return g.Wait()
}

// tick queries price from ojo and broadcasts wasm tx with prices to the wasm contract periodically.
func (r *Relayer) tick(ctx context.Context) error {
	r.logger.Debug().Msg("executing relayer tick")

	blockHeight, err := r.relayerClient.ChainHeight.GetBlockHeight()
	if err != nil {
		return err
	}
	if blockHeight < 1 {
		return fmt.Errorf("expected positive block height")
	}

	blockTimestamp, err := r.relayerClient.ChainHeight.GetBlockTime()
	if err != nil {
		return err
	}

	if blockTimestamp < 1 {
		return fmt.Errorf("expected positive blocktimestamp")
	}

	var postMedian bool
	if r.medianDuration > 0 {
		postMedian = r.requestID%uint64(r.medianDuration) == 0
	}

	if err := r.setDenomPrices(ctx, postMedian); err != nil {
		return err
	}

	nextBlockHeight := blockHeight + 1

	forceRelay := r.missedCounter >= r.missedThreshold

	// set the next resolve time for price feeds on evm contract
	nextBlockTime := blockTimestamp + uint64(r.resolveDuration.Nanoseconds())
	exchangeMsg := r.genRateMsgs(r.requestID, nextBlockTime)
	if err != nil {
		return err
	}

	deviationMsg := r.genDeviationsMsg(r.deviationRequestID, nextBlockTime)
	if err != nil {
		return err
	}

	var medianMsg []client.PriceFeedMedianData
	if postMedian {
		resolveTime := time.Duration(r.resolveDuration.Nanoseconds() * r.medianDuration)
		nextMedianBlockTime := blockTimestamp + uint64(resolveTime.Seconds())
		medianMsg = r.genMedianMsg(r.medianRequestID, nextMedianBlockTime)
		if err != nil {
			return err
		}
	}

	logs := r.logger.Info()
	logs.Str("contract address", r.contractAddress).
		Str("relayer address", r.relayerClient.RelayerAddress.String()).
		Uint64("block timestamp", blockTimestamp).
		Bool("median posted", postMedian).
		Uint64("request id", r.requestID).
		Uint64("deviation request id", r.deviationRequestID)

	if postMedian {
		logs.Uint64("median request id", r.medianRequestID)
	}

	logs.Msg("broadcasting execute to contract")

	if err := r.relayerClient.BroadcastTx(nextBlockHeight, 1000000, exchangeMsg, deviationMsg, medianMsg, forceRelay); err != nil {
		r.missedCounter += 1
		return err
	}

	// reset missed counter if force relay is successful
	if forceRelay {
		r.missedCounter = 0
	}

	// increment request id to be stored in contracts
	r.requestID += 1
	r.deviationRequestID += 1
	if postMedian {
		r.medianRequestID += 1
	}

	return nil
}

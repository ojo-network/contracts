package relayer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	oracletypes "github.com/ojo-network/ojo/x/oracle/types"

	psync "github.com/ojo-network/cw-relayer/pkg/sync"
	"github.com/ojo-network/cw-relayer/relayer/client"
	scrttypes "github.com/ojo-network/cw-relayer/relayer/dep"
	scrtutils "github.com/ojo-network/cw-relayer/relayer/dep/utils"
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

	contractAddr types.AccAddress
	codeHash     string

	requestID          uint64
	medianRequestID    uint64
	deviationRequestID uint64

	exchangeRates        types.DecCoins
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

	iopubkey []byte

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
	queryRPCS []string,
	codeHash string,
	resolveDuration time.Duration,
	queryTimeout time.Duration,
	requestID uint64,
	medianRequestID uint64,
	deviationRequestID uint64,
	medianDuration int64,
	deviationDuration int64,
	ignoreMedianErrors bool,
	config AutoRestartConfig,
	event chan struct{},
) *Relayer {

	contractAddr, err := types.AccAddressFromBech32(contractAddress)
	if err != nil {
		panic(err)
	}

	return &Relayer{
		queryRPCS:          queryRPCS,
		logger:             logger.With().Str("module", "relayer").Logger(),
		relayerClient:      oc,
		contractAddress:    contractAddress,
		contractAddr:       contractAddr,
		missedThreshold:    missedThreshold,
		timeoutHeight:      timeoutHeight,
		closer:             psync.NewCloser(),
		codeHash:           codeHash,
		medianDuration:     medianDuration,
		deviationDuration:  deviationDuration,
		ignoreMedianErrors: ignoreMedianErrors,
		resolveDuration:    resolveDuration,
		requestID:          requestID,
		medianRequestID:    medianRequestID,
		deviationRequestID: deviationRequestID,
		maxQueryRetries:    maxQueryRetries,
		queryTimeout:       queryTimeout,
		config:             config,
		event:              event,
	}
}

func (r *Relayer) Start(ctx context.Context) error {
	clientCtx, err := r.relayerClient.CreateClientContext()
	if err != nil {
		return err
	}

	// iopubkey for encryption
	var iopubkey []byte
	iopubkey, err = scrtutils.GetConsensusIoPubKey(scrtutils.WASMContext{CLIContext: clientCtx})
	if err != nil {
		return err
	}

	r.iopubkey = iopubkey

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
	queryMsgs, err := genRestartQueries(r.contractAddress, r.config.Denom)
	if err != nil {
		return err
	}

	responses, err := r.relayerClient.BroadcastContractQuery(
		ctx,
		r.iopubkey,
		r.codeHash,
		r.contractAddress,
		r.queryTimeout,
		queryMsgs...,
	)
	if err != nil {
		return err
	}

	for _, response := range responses {
		if len(response.QueryResponse.Data) != 0 {
			var resp map[string]interface{}
			err := json.Unmarshal(response.QueryResponse.Data, &resp)
			if err != nil {
				return err
			}

			id, err := strconv.ParseInt(resp["request_id"].(string), 10, 64)
			if err != nil {
				return err
			}

			// increment request id for relay
			requestID := uint64(id) + 1
			switch response.QueryType {
			case int(QueryRateMsg):
				r.requestID = requestID
			case int(QueryMedianRateMsg):
				r.medianRequestID = requestID
			case int(QueryDeviationRateMsg):
				r.deviationRequestID = requestID
			}
		}
	}

	return nil
}

func (r *Relayer) setDenomPrices(ctx context.Context, postMedian, postDeviation bool) error {
	if r.queryRetries > r.maxQueryRetries {
		r.queryRetries = 0
		return noRates
	}

	grpcConn, err := grpc.Dial(
		r.queryRPCS[r.index],
		// the Cosmos SDK doesn't support any transport security mechanism
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialerFunc),
	)

	// retry or switch rpc
	if err != nil {
		r.increment()
		return r.setDenomPrices(ctx, postMedian, postDeviation)
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
		return r.setDenomPrices(ctx, postMedian, postDeviation)
	}

	r.exchangeRates = queryResponse.ExchangeRates

	var mu sync.Mutex
	g, _ := errgroup.WithContext(ctx)

	if postDeviation {
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
	}

	if postMedian {
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
	}

	return g.Wait()
}

// tick queries price from ojo and broadcasts wasm tx with prices to the wasm contract periodically.
func (r *Relayer) tick(ctx context.Context) error {
	r.logger.Debug().Msg("executing relayer tick")

	blockHeight, err := r.relayerClient.ChainHeight.GetChainHeight()
	if err != nil {
		return err
	}
	if blockHeight < 1 {
		return fmt.Errorf("expected positive block height")
	}

	blockTimestamp, err := r.relayerClient.ChainHeight.GetChainTimestamp()
	if err != nil {
		return err
	}

	if blockTimestamp.Unix() < 1 {
		return fmt.Errorf("expected positive blocktimestamp")
	}

	var postMedian bool
	if r.medianDuration > 0 {
		postMedian = r.requestID%uint64(r.medianDuration) == 0
	}

	var postDeviation bool
	if r.deviationDuration > 0 {
		postDeviation = r.requestID%uint64(r.deviationDuration) == 0
	}

	err = r.setDenomPrices(ctx, postMedian, postDeviation)
	switch err {
	case nil:
		break
	case noMedians, noDeviations:
		if !r.ignoreMedianErrors {
			return err
		}

		// as median and deviation are not properly set, do not push prices to contract
		postMedian = false
		postDeviation = false
	default:
		return err
	}

	nextBlockHeight := blockHeight + 1

	forceRelay := r.missedCounter >= r.missedThreshold

	// set the next resolve time for price feeds on wasm contract
	nextBlockTime := blockTimestamp.Add(r.resolveDuration).Unix()
	rateMsg, err := generateContractRelayMsg(forceRelay, RelayRate, r.requestID, nextBlockTime, r.exchangeRates)
	if err != nil {
		return err
	}

	dataMsgs := [][]byte{rateMsg}

	logs := r.logger.Info()
	logs.Str("contract address", r.contractAddress).
		Str("relayer address", r.relayerClient.RelayerAddrString).
		Str("block timestamp", blockTimestamp.String()).
		Bool("median posted", postMedian).
		Bool("deviation posted", postDeviation).
		Uint64("request id", r.requestID)

	if postDeviation {
		resolveTime := time.Duration(r.resolveDuration.Nanoseconds() * r.deviationDuration)
		nextDeviationBlockTime := blockTimestamp.Add(resolveTime).Unix()
		deviationMsg, err := generateContractRelayMsg(
			forceRelay,
			RelayHistoricalDeviation,
			r.deviationRequestID,
			nextDeviationBlockTime,
			r.historicalDeviations,
		)
		if err != nil {
			return err
		}

		dataMsgs = append(dataMsgs, deviationMsg)
		logs.Uint64("deviation request id", r.deviationRequestID)
	}

	if postMedian {
		resolveTime := time.Duration(r.resolveDuration.Nanoseconds() * r.medianDuration)
		nextMedianBlockTime := blockTimestamp.Add(resolveTime).Unix()
		medianMsg, err := generateContractRelayMsg(
			forceRelay,
			RelayHistoricalMedian,
			r.medianRequestID,
			nextMedianBlockTime,
			r.historicalMedians,
		)
		if err != nil {
			return err
		}

		dataMsgs = append(dataMsgs, medianMsg)
		logs.Uint64("median request id", r.medianRequestID)
	}

	clientCtx, wasmCtx, err := r.relayerClient.CreateClientAndScrtWasmCtx()
	if err != nil {
		return err
	}

	executeMsgs := make([]types.Msg, len(dataMsgs))
	for i, msg := range dataMsgs {
		execMsg := scrttypes.NewSecretMsg([]byte(r.codeHash), msg)
		encryptedMsg, err := wasmCtx.Encrypt(r.iopubkey, execMsg.Serialize())
		if err != nil {
			return err
		}

		executeMsg := scrttypes.MsgExecuteContract{
			Sender:   r.relayerClient.RelayerAddr,
			Contract: r.contractAddr,
			Msg:      encryptedMsg,
		}

		executeMsgs[i] = &executeMsg
	}

	logs.Msg("broadcasting execute to contract")
	if err := r.relayerClient.BroadcastTx(*clientCtx, nextBlockHeight, r.timeoutHeight, executeMsgs...); err != nil {
		r.missedCounter += 1
		return err
	}

	// reset missed counter if force relay is successful
	if forceRelay {
		r.missedCounter = 0
	}

	// increment request id to be stored in contracts
	r.requestID += 1
	if postMedian {
		r.medianRequestID += 1
	}

	if postDeviation {
		r.deviationRequestID += 1
	}

	return nil
}

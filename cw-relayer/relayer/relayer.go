package relayer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/cosmos/cosmos-sdk/types"
	oracletypes "github.com/ojo-network/ojo/x/oracle/types"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ojo-network/cw-relayer/pkg/sync"
	"github.com/ojo-network/cw-relayer/relayer/client"
)

var (
	// RateFactor is used to convert ojo prices to contract-compatible values.
	RateFactor = types.NewDec(10).Power(9)
)

// Relayer defines a structure that queries prices from ojo and publishes prices to wasm contract.
type Relayer struct {
	logger zerolog.Logger
	closer *sync.Closer

	relayerClient   client.RelayerClient
	queryRPCS       []string
	contractAddress string
	requestID       uint64
	medianRequestID uint64

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

	event chan struct{}
}

// New returns an instance of the relayer.
func New(
	logger zerolog.Logger,
	oc client.RelayerClient,
	contractAddress string,
	timeoutHeight int64,
	missedThreshold int64,
	queryRPCS []string,
	event chan struct{},
	medianDuration int64,
	resolveDuration time.Duration,
	queryTimeout time.Duration,
	requestID uint64,
	medianRequestID uint64,
) *Relayer {
	return &Relayer{
		queryRPCS:       queryRPCS,
		logger:          logger.With().Str("module", "relayer").Logger(),
		relayerClient:   oc,
		contractAddress: contractAddress,
		missedThreshold: missedThreshold,
		timeoutHeight:   timeoutHeight,
		queryTimeout:    queryTimeout,
		medianDuration:  medianDuration,
		resolveDuration: resolveDuration,
		requestID:       requestID,
		medianRequestID: medianRequestID,
		closer:          sync.NewCloser(),
		event:           event,
	}
}

func (r *Relayer) Start(ctx context.Context) error {
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

func (r *Relayer) setDenomPrices(ctx context.Context, postMedian bool) error {
	g, ctx := errgroup.WithContext(ctx)
	grpcConn := &grpc.ClientConn{}
	var err error
	grpcConn, err = grpc.Dial(
		r.queryRPCS[0],
		// the Cosmos SDK doesn't support any transport security mechanism
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialerFunc),
	)

	// err if no rpc has successfully connected
	if err != nil {
		return err
	}

	defer grpcConn.Close()

	queryClient := oracletypes.NewQueryClient(grpcConn)

	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	g.Go(func() error {
		queryResponse, err := queryClient.ExchangeRates(ctx, &oracletypes.QueryExchangeRates{})
		if err != nil {
			return err
		}

		if queryResponse.ExchangeRates.Empty() {
			return fmt.Errorf("exchange rates empty")
		}

		r.exchangeRates = queryResponse.ExchangeRates
		return nil
	})

	g.Go(func() error {
		deviationsQueryResponse, err := queryClient.MedianDeviations(ctx, &oracletypes.QueryMedianDeviations{})
		if err != nil {
			return err
		}

		if deviationsQueryResponse.MedianDeviations.Empty() {
			return fmt.Errorf("median deviations empty")
		}

		r.historicalDeviations = deviationsQueryResponse.MedianDeviations
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

			r.historicalMedians = medianQueryResponse.Medians
			return nil
		})
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

	if err := r.setDenomPrices(ctx, postMedian); err != nil {
		return err
	}

	nextBlockHeight := blockHeight + 1

	forceRelay := r.missedCounter >= r.missedThreshold

	// set the next resolve time for price feeds on wasm contract
	nextBlockTime := blockTimestamp.Add(r.resolveDuration).Unix()
	exchangeMsg, err := genRateMsgData(forceRelay, RelayRate, r.requestID, nextBlockTime, r.exchangeRates)
	if err != nil {
		return err
	}

	deviationMsg, err := genRateMsgData(forceRelay, RelayHistoricalDeviation, r.requestID, nextBlockTime, r.historicalDeviations)
	if err != nil {
		return err
	}

	var msgs []types.Msg
	msgs = append(msgs, r.genWasmMsg(exchangeMsg), r.genWasmMsg(deviationMsg))

	if postMedian {
		resolveTime := time.Duration(r.resolveDuration.Nanoseconds() * r.medianDuration)
		nextMedianBlockTime := blockTimestamp.Add(resolveTime).Unix()
		medianMsg, err := genRateMsgData(forceRelay, RelayHistoricalMedian, r.medianRequestID, nextMedianBlockTime, r.historicalMedians)
		if err != nil {
			return err
		}

		msgs = append(msgs, r.genWasmMsg(medianMsg))
	}

	logs := r.logger.Info()
	logs.Str("contract address", r.contractAddress).
		Str("relayer address", r.relayerClient.RelayerAddrString).
		Str("block timestamp", blockTimestamp.String()).
		Bool("median posted", postMedian).
		Uint64("request id", r.requestID)

	if postMedian {
		logs.Uint64("median request id", r.medianRequestID)
	}

	logs.Msg("broadcasting execute to contract")

	if err := r.relayerClient.BroadcastTx(nextBlockHeight, r.timeoutHeight, msgs...); err != nil {
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

	return nil
}

func (r *Relayer) genWasmMsg(msgData []byte) *wasmtypes.MsgExecuteContract {
	return &wasmtypes.MsgExecuteContract{
		Sender:   r.relayerClient.RelayerAddrString,
		Contract: r.contractAddress,
		Msg:      msgData,
		Funds:    nil,
	}
}

func genRateMsgData(forceRelay bool, msgType MsgType, requestID uint64, resolveTime int64, rates types.DecCoins) (msgData []byte, err error) {
	msg := Msg{
		SymbolRates: nil,
		ResolveTime: resolveTime,
		RequestID:   requestID,
	}

	if msgType != RelayHistoricalMedian {
		for _, rate := range rates {
			msg.SymbolRates = append(msg.SymbolRates, [2]interface{}{rate.Denom, rate.Amount.Mul(RateFactor).TruncateInt().String()})
		}
	}

	switch msgType {
	case RelayRate:
		if forceRelay {
			msgData, err = json.Marshal(MsgForceRelay{Relay: msg})
		} else {
			msgData, err = json.Marshal(MsgRelay{Relay: msg})
		}
	case RelayHistoricalMedian:
		// collect denom's medians
		medianRates := map[string][]string{}
		for _, rate := range rates {
			medianRates[rate.Denom] = append(medianRates[rate.Denom], rate.Amount.Mul(RateFactor).TruncateInt().String())
		}

		for denom, medians := range medianRates {
			msg.SymbolRates = append(msg.SymbolRates, [2]interface{}{denom, medians})
		}

		if forceRelay {
			msgData, err = json.Marshal(MsgForceRelayHistoricalMedian{Relay: msg})
		} else {
			msgData, err = json.Marshal(MsgRelayHistoricalMedian{Relay: msg})
		}
	case RelayHistoricalDeviation:
		if forceRelay {
			msgData, err = json.Marshal(MsgForceRelayHistoricalDeviation{Relay: msg})
		} else {
			msgData, err = json.Marshal(MsgRelayHistoricalDeviation{Relay: msg})
		}
	}

	return
}

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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ojo-network/cw-relayer/pkg/sync"
	"github.com/ojo-network/cw-relayer/relayer/client"
)

const (
	tickerSleep = 1000 * time.Millisecond
)

var (
	// Factor is multiplied to price to convert them into correct amount for contracts
	Factor = types.NewDec(10).Power(9)
)

// Relayer defines a structure that queries prices from ojo and publishes prices to wasm contract.
type Relayer struct {
	logger zerolog.Logger
	closer *sync.Closer

	relayerClient   client.RelayerClient
	exchangeRates   types.DecCoins
	queryRPC        string
	contractAddress string
	requestID       uint64
	timeoutHeight   int64

	// if missedCounter >= missedThreshold, force relay prices (bypasses timing restrictions)
	missedCounter   int64
	missedThreshold int64
}

// New returns an instance of the relayer.
func New(
	logger zerolog.Logger,
	oc client.RelayerClient,
	contractAddress string,
	timeoutHeight int64,
	missedThreshold int64,
	queryRPC string,
) *Relayer {
	return &Relayer{
		queryRPC:        queryRPC,
		logger:          logger.With().Str("module", "relayer").Logger(),
		relayerClient:   oc,
		contractAddress: contractAddress,
		missedThreshold: missedThreshold,
		timeoutHeight:   timeoutHeight,
		closer:          sync.NewCloser(),
	}
}

func (r *Relayer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			r.closer.Close()

		default:
			r.logger.Debug().Msg("starting relayer tick")

			startTime := time.Now()
			if err := r.tick(ctx); err != nil {
				telemetry.IncrCounter(1, "failure", "tick")
				r.logger.Err(err).Msg("relayer tick failed")
			}

			telemetry.MeasureSince(startTime, "runtime", "tick")
			telemetry.IncrCounter(1, "new", "tick")

			time.Sleep(tickerSleep)
		}
	}
}

// Stop stops the relayer process and waits for it to gracefully exit.
func (r *Relayer) Stop() {
	r.closer.Close()
	<-r.closer.Done()
}

func (r *Relayer) setActiveDenomPrices(ctx context.Context) error {
	grpcConn, err := grpc.Dial(
		r.queryRPC,
		// the Cosmos SDK doesn't support any transport security mechanism
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialerFunc),
	)
	if err != nil {
		return err
	}

	defer grpcConn.Close()

	queryClient := oracletypes.NewQueryClient(grpcConn)

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	queryResponse, err := queryClient.ExchangeRates(ctx, &oracletypes.QueryExchangeRates{})
	if err != nil {
		return err
	}

	r.exchangeRates = queryResponse.ExchangeRates
	return nil
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

	if err := r.setActiveDenomPrices(ctx); err != nil {
		return err
	}
	nextBlockHeight := blockHeight + 1

	forceRelay := r.missedCounter >= r.missedThreshold
	if forceRelay {
		r.missedCounter = 0
	}

	// set the next resolve time for price feeds on wasm contract
	nextBlockTime := blockTimestamp.Unix() + int64(tickerSleep.Seconds())
	msg, err := generateContractRelayMsg(forceRelay, r.requestID, nextBlockTime, r.exchangeRates)
	if err != nil {
		return err
	}

	// increment request id to be stored in contracts
	r.requestID += 1

	executeMsg := &wasmtypes.MsgExecuteContract{
		Sender:   r.relayerClient.RelayerAddrString,
		Contract: r.contractAddress,
		Msg:      msg,
		Funds:    nil,
	}

	r.logger.Info().
		Str("Contract Address", executeMsg.Contract).
		Str("relayer addr", executeMsg.Sender).
		Str("block timestamp", blockTimestamp.String()).
		Msg("broadcasting execute to contract")

	if err := r.relayerClient.BroadcastTx(nextBlockHeight, r.timeoutHeight, executeMsg); err != nil {
		return err
	}

	return nil
}

func generateContractRelayMsg(forceRelay bool, requestID uint64, resolveTime int64, exchangeRates types.DecCoins) ([]byte, error) {
	msg := Msg{
		SymbolRates: nil,
		ResolveTime: resolveTime,
		RequestID:   requestID,
	}

	for _, rate := range exchangeRates {
		msg.SymbolRates = append(msg.SymbolRates, [2]string{rate.Denom, rate.Amount.Mul(Factor).TruncateInt().String()})
	}

	if forceRelay {
		msgData, err := json.Marshal(MsgForceRelay{Relay: msg})
		return msgData, err
	}

	msgData, err := json.Marshal(MsgRelay{Relay: msg})
	return msgData, err
}

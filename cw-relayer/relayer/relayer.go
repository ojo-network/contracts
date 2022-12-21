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

type Relayer struct {
	logger zerolog.Logger
	closer *sync.Closer

	relayerClient   client.RelayerClient
	exchangeRates   types.DecCoins
	queryRPC        string
	contractAddress string
	requestID       uint64
	timeoutHeight   int64
	missedCounter   int64
	missedThreshold int64
}

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

func (o *Relayer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			o.closer.Close()

		default:
			o.logger.Debug().Msg("starting relayer tick")

			startTime := time.Now()
			if err := o.tick(ctx); err != nil {
				telemetry.IncrCounter(1, "failure", "tick")
				o.logger.Err(err).Msg("relayer tick failed")
			}

			telemetry.MeasureSince(startTime, "runtime", "tick")
			telemetry.IncrCounter(1, "new", "tick")

			time.Sleep(tickerSleep)
		}
	}
}

// Stop stops the relayer process and waits for it to gracefully exit.
func (o *Relayer) Stop() {
	o.closer.Close()
	<-o.closer.Done()
}

func (o *Relayer) setActiveDenomPrices(ctx context.Context) error {
	grpcConn, err := grpc.Dial(
		o.queryRPC,
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

	o.exchangeRates = queryResponse.ExchangeRates
	return nil
}

func (o *Relayer) tick(ctx context.Context) error {
	o.logger.Debug().Msg("executing relayer tick")

	blockHeight, err := o.relayerClient.ChainHeight.GetChainHeight()
	if err != nil {
		return err
	}
	if blockHeight < 1 {
		return fmt.Errorf("expected positive block height")
	}

	blockTimestamp, err := o.relayerClient.ChainHeight.GetChainTimestamp()
	if err != nil {
		return err
	}

	if blockTimestamp.Unix() < 1 {
		return fmt.Errorf("expected positive blocktimestamp")
	}

	if err := o.setActiveDenomPrices(ctx); err != nil {
		return err
	}
	nextBlockHeight := blockHeight + 1

	forceRelay := o.missedCounter >= o.missedThreshold
	if forceRelay {
		o.missedCounter = 0
	}

	// set the next resolve time for price feeds on wasm contract
	nextBlockTime := blockTimestamp.Unix() + int64(tickerSleep.Seconds())
	msg, err := generateContractRelayMsg(forceRelay, o.requestID, nextBlockTime, o.exchangeRates)
	if err != nil {
		return err
	}

	o.requestID += 1
	executeMsg := &wasmtypes.MsgExecuteContract{
		Sender:   o.relayerClient.RelayerAddrString,
		Contract: o.contractAddress,
		Msg:      msg,
		Funds:    nil,
	}

	o.logger.Info().
		Str("Contract Address", executeMsg.Contract).
		Str("relayer addr", executeMsg.Sender).
		Str("block timestamp", blockTimestamp.String()).
		Msg("broadcasting execute to contract")

	if err := o.relayerClient.BroadcastTx(nextBlockHeight, o.timeoutHeight, executeMsg); err != nil {
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

	factor := types.NewDec(10).Power(18)
	for _, rate := range exchangeRates {
		msg.SymbolRates = append(msg.SymbolRates, [2]string{rate.Denom, rate.Amount.Mul(factor).TruncateInt().String()})
	}

	if forceRelay {
		msgData, err := json.Marshal(MsgForceRelay{Relay: msg})
		return msgData, err
	}

	msgData, err := json.Marshal(MsgRelay{Relay: msg})
	return msgData, err
}

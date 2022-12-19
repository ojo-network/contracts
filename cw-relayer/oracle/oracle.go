package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/cosmos/cosmos-sdk/types"
	oracletypes "github.com/ojo-network/ojo/x/oracle/types"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ojo-network/cw-relayer/config"
	"github.com/ojo-network/cw-relayer/oracle/client"
	psync "github.com/ojo-network/cw-relayer/pkg/sync"
)

const (
	tickerSleep = 1000 * time.Millisecond
)

type Oracle struct {
	logger        zerolog.Logger
	closer        *psync.Closer
	oracleClient  client.OracleClient
	config        config.Config
	ExchangeRates types.DecCoins
	TargetRPC     string
	pricesMutex   sync.RWMutex
	prices        map[string]types.Dec
	requestID     uint64
	timeoutHeight int64
}

func New(logger zerolog.Logger,
	oc client.OracleClient,
	config config.Config,
	providerTimeout time.Duration,
	targetRPC string) *Oracle {
	return &Oracle{
		TargetRPC:    targetRPC,
		logger:       logger.With().Str("module", "oracle").Logger(),
		oracleClient: oc,
		config:       config,
	}
}
func (o *Oracle) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			o.closer.Close()

		default:
			o.logger.Debug().Msg("starting oracle tick")

			startTime := time.Now()

			if err := o.tick(ctx); err != nil {
				telemetry.IncrCounter(1, "failure", "tick")
				o.logger.Err(err).Msg("oracle tick failed")
			}

			telemetry.MeasureSince(startTime, "runtime", "tick")
			telemetry.IncrCounter(1, "new", "tick")

			time.Sleep(tickerSleep)
		}
	}
}

// Stop stops the oracle process and waits for it to gracefully exit.
func (o *Oracle) Stop() {
	o.closer.Close()
	<-o.closer.Done()
}

func (o *Oracle) SetActiveDenomPrices(ctx context.Context) error {
	grpcConn, err := grpc.Dial(
		o.TargetRPC,
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

	o.ExchangeRates = queryResponse.ExchangeRates

	return nil
}

func (o *Oracle) tick(ctx context.Context) error {
	o.logger.Debug().Msg("executing oracle tick")

	blockHeight, err := o.oracleClient.ChainHeight.GetChainHeight()
	if err != nil {
		return err
	}
	if blockHeight < 1 {
		return fmt.Errorf("expected positive block height")
	}

	blockTimestamp, err := o.oracleClient.ChainHeight.GetChainTimestamp()
	if err != nil {
		return err
	}
	if blockTimestamp.Unix() < 1 {
		return fmt.Errorf("expected positive blocktimestamp")
	}

	if err := o.SetActiveDenomPrices(ctx); err != nil {
		return err
	}
	nextBlockHeight := blockHeight + 1

	msg, err := generateContractMsg(o.requestID, blockTimestamp.Unix()+int64(tickerSleep.Seconds()), o.ExchangeRates)
	if err != nil {
		return err
	}

	o.requestID += 1

	executeMsg := &wasmtypes.MsgExecuteContract{
		Sender:   o.oracleClient.OracleAddrString,
		Contract: o.config.ContractAddress,
		Msg:      msg,
		Funds:    nil,
	}

	o.logger.Info().
		Str("Contract Address", executeMsg.Contract).
		Str("oracle addr", executeMsg.Sender).
		Msg("broadcasting execute contract")
	if err := o.oracleClient.BroadcastTx(nextBlockHeight, o.timeoutHeight, executeMsg); err != nil {
		return err
	}

	return nil
}

func generateContractMsg(requestID uint64, resolveTime int64, exchangeRates types.DecCoins) ([]byte, error) {
	msg := config.Relay{
		SymbolRates: nil,
		ResolveTime: resolveTime,
		RequestID:   requestID,
	}

	for _, rate := range exchangeRates {
		msg.SymbolRates = append(msg.SymbolRates, [2]string{rate.Denom, rate.Amount.String()})
	}

	msgData, err := json.Marshal(msg)
	return msgData, err
}

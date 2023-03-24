package relayer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	oracletypes "github.com/ojo-network/ojo/x/oracle/types"

	"github.com/ojo-network/cw-relayer/pkg/sync"
	"github.com/ojo-network/cw-relayer/relayer/client"
	scrttypes "github.com/ojo-network/cw-relayer/relayer/dep"
	scrtutils "github.com/ojo-network/cw-relayer/relayer/dep/utils"
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
	exchangeRates   types.DecCoins
	queryRPCS       []string
	contractAddress string
	requestID       uint64
	contractAddr    types.AccAddress
	codeHash        string

	resolveDuration time.Duration
	queryTimeout    time.Duration

	// if missedCounter >= missedThreshold, force relay prices (bypasses timing restrictions)
	missedCounter   int64
	missedThreshold int64
	timeoutHeight   int64
	maxQueryRetries int64
	queryRetries    int64
	index           int

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
	event chan struct{},
) *Relayer {

	contractAddr, err := types.AccAddressFromBech32(contractAddress)
	if err != nil {
		panic(err)
	}

	return &Relayer{
		queryRPCS:       queryRPCS,
		logger:          logger.With().Str("module", "relayer").Logger(),
		relayerClient:   oc,
		contractAddress: contractAddress,
		missedThreshold: missedThreshold,
		timeoutHeight:   timeoutHeight,
		closer:          sync.NewCloser(),
		contractAddr:    contractAddr,
		codeHash:        codeHash,
		resolveDuration: resolveDuration,
		requestID:       requestID,
		maxQueryRetries: maxQueryRetries,
		queryTimeout:    queryTimeout,
		event:           event,
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

func (r *Relayer) restart(ctx context.Context) error {
	queryMsgs, err := restartQuery(r.contractAddress, r.config.Denom)
	if err != nil {
		return err
	}

	responses, err := r.relayerClient.BroadcastContractQuery(ctx, r.queryTimeout, queryMsgs)
	if err != nil {
		return err
	}

	for _, response := range responses {
		if len(response.QueryResponse.Data) != 0 {
			var resp map[string]interface{}
			err := json.Unmarshal(response.QueryResponse.Data, &resp)
			if err != nil {
				return nil
			}

			id, err := strconv.ParseInt(resp["request_id"].(string), 10, 64)
			if err != nil {
				return err
			}

			requestID := uint64(id)
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
		grpc.WithContextDialer(dialerFunc),
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

	if err := r.setDenomPrices(ctx); err != nil {
		return err
	}
	nextBlockHeight := blockHeight + 1

	forceRelay := r.missedCounter >= r.missedThreshold

	// set the next resolve time for price feeds on wasm contract
	nextBlockTime := blockTimestamp.Add(r.resolveDuration).Unix()
	msg, err := generateContractRelayMsg(forceRelay, r.requestID, nextBlockTime, r.exchangeRates)
	if err != nil {
		return err
	}

	execMsg := scrttypes.NewSecretMsg([]byte(r.codeHash), msg)
	clientCtx, err := r.relayerClient.CreateClientContext()
	if err != nil {
		return err
	}

	wasmCtx := scrtutils.WASMContext{CLIContext: clientCtx}
	encryptedMsg, err := wasmCtx.Encrypt(r.iopubkey, execMsg.Serialize())
	if err != nil {
		return err
	}

	executeMsg := &scrttypes.MsgExecuteContract{
		Sender:   r.relayerClient.RelayerAddr,
		Contract: r.contractAddr,
		Msg:      encryptedMsg,
	}

	r.logger.Info().
		Str("Contract Address", r.contractAddress).
		Str("Relayer Address", r.relayerClient.RelayerAddrString).
		Str("block timestamp", blockTimestamp.String()).
		Msg("broadcasting execute to contract")

	if err := r.relayerClient.BroadcastTx(clientCtx, nextBlockHeight, r.timeoutHeight, executeMsg); err != nil {
		r.missedCounter += 1
		return err
	}

	// reset missed counter if force relay is successful
	if forceRelay {
		r.missedCounter = 0
	}

	// increment request id to be stored in contracts
	r.requestID += 1

	return nil
}

func generateContractRelayMsg(forceRelay bool, requestID uint64, resolveTime int64, exchangeRates types.DecCoins) ([]byte, error) {
	msg := Msg{
		SymbolRates: nil,
		ResolveTime: resolveTime,
		RequestID:   requestID,
	}

	for _, rate := range exchangeRates {
		msg.SymbolRates = append(msg.SymbolRates, [2]string{rate.Denom, rate.Amount.Mul(RateFactor).TruncateInt().String()})
	}

	if forceRelay {
		msgData, err := json.Marshal(MsgForceRelay{Relay: msg})
		return msgData, err
	}

	msgData, err := json.Marshal(MsgRelay{Relay: msg})

	return msgData, err
}

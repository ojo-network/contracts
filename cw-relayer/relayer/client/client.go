package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/credentials/insecure"

	wasmparams "github.com/CosmWasm/wasmd/app/params"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/rs/zerolog"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	tmjsonclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"google.golang.org/grpc"
)

type (
	// RelayerClient defines a structure that interfaces with the smart-contract-enabled chain.
	RelayerClient struct {
		logger            zerolog.Logger
		ChainID           string
		KeyringBackend    string
		KeyringDir        string
		KeyringPass       string
		TMRPC             []string
		QueryRpc          []string
		RPCTimeout        time.Duration
		RelayerAddr       sdk.AccAddress
		RelayerAddrString string
		Encoding          wasmparams.EncodingConfig
		GasPrices         string
		GasAdjustment     float64
		KeyringPassphrase string
		ChainHeight       *ChainHeight
	}

	passReader struct {
		pass string
		buf  *bytes.Buffer
	}
)

type SmartQuery struct {
	QueryType int
	QueryMsg  wasmtypes.QuerySmartContractStateRequest
}

type QueryResponse struct {
	QueryType     int
	QueryResponse wasmtypes.QuerySmartContractStateResponse
}

func NewRelayerClient(
	ctx context.Context,
	logger zerolog.Logger,
	chainID string,
	keyringBackend string,
	keyringDir string,
	keyringPass string,
	tmRPC []string,
	queryEndpoint []string,
	rpcTimeout time.Duration,
	RelayerAddrString string,
	accPrefix string,
	gasAdjustment float64,
	GasPrices string,
) (RelayerClient, error) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(accPrefix, accPrefix+sdk.PrefixPublic)
	config.Seal()

	RelayerAddr, err := sdk.AccAddressFromBech32(RelayerAddrString)
	if err != nil {
		return RelayerClient{}, err
	}

	relayerClient := RelayerClient{
		logger:            logger.With().Str("module", "relayer_client").Logger(),
		ChainID:           chainID,
		KeyringBackend:    keyringBackend,
		KeyringDir:        keyringDir,
		KeyringPass:       keyringPass,
		TMRPC:             tmRPC,
		RPCTimeout:        rpcTimeout,
		RelayerAddr:       RelayerAddr,
		RelayerAddrString: RelayerAddrString,
		Encoding:          MakeEncodingConfig(),
		GasAdjustment:     gasAdjustment,
		GasPrices:         GasPrices,
		QueryRpc:          queryEndpoint,
	}

	clientCtx, err := relayerClient.CreateClientContext()
	if err != nil {
		return RelayerClient{}, err
	}

	blockHeight, err := rpc.GetChainHeight(clientCtx)
	if err != nil {
		return RelayerClient{}, err
	}

	blockTime, err := GetChainTimestamp(clientCtx)
	if err != nil {
		return RelayerClient{}, err
	}

	chainHeight, err := NewChainHeight(
		ctx,
		clientCtx.Client,
		relayerClient.logger,
		blockHeight,
		blockTime,
	)
	if err != nil {
		return RelayerClient{}, err
	}
	relayerClient.ChainHeight = chainHeight

	return relayerClient, nil
}

func newPassReader(pass string) io.Reader {
	return &passReader{
		pass: pass,
		buf:  new(bytes.Buffer),
	}
}

func (r *passReader) Read(p []byte) (n int, err error) {
	n, err = r.buf.Read(p)
	if err == io.EOF || n == 0 {
		r.buf.WriteString(r.pass + "\n")

		n, err = r.buf.Read(p)
	}

	return n, err
}

// BroadcastTx attempts to broadcast a signed transaction. If it fails, a few re-attempts
// will be made until the transaction succeeds or ultimately times out or fails.
func (oc RelayerClient) BroadcastTx(timeoutHeight int64, msgs ...sdk.Msg) error {
	nextBlockHeight, err := oc.ChainHeight.GetChainHeight()
	if err != nil {
		return nil
	}

	maxBlockHeight := nextBlockHeight + timeoutHeight
	lastCheckHeight := nextBlockHeight - 1

	clientCtx, err := oc.CreateClientContext()
	if err != nil {
		return err
	}

	factory, err := oc.CreateTxFactory(&clientCtx)
	if err != nil {
		return err
	}

	// re-try tx until timeout
	for lastCheckHeight < maxBlockHeight {
		latestBlockHeight, err := oc.ChainHeight.GetChainHeight()
		if err != nil {
			return err
		}

		if latestBlockHeight <= lastCheckHeight {
			continue
		}

		// set last check height to latest block height
		lastCheckHeight = latestBlockHeight
		resp, err := BroadcastTx(clientCtx, factory, msgs...)
		if resp != nil && resp.Code != 0 {
			telemetry.IncrCounter(1, "failure", "tx", "code")
			oc.logger.Error().Msg(resp.String())
			err = fmt.Errorf("invalid response code from tx: %d", resp.Code)
		}

		if err != nil {
			var (
				code uint32
				hash string
			)
			if resp != nil {
				code = resp.Code
				hash = resp.TxHash
			}

			oc.logger.Debug().
				Err(err).
				Int64("max_height", maxBlockHeight).
				Int64("last_check_height", lastCheckHeight).
				Str("tx_hash", hash).
				Uint32("tx_code", code).
				Msg("failed to broadcast tx; retrying...")

			time.Sleep(time.Second * 1)
			continue
		}

		oc.logger.Info().
			Uint32("tx_code", resp.Code).
			Str("tx_hash", resp.TxHash).
			Int64("tx_height", resp.Height).
			Msg("successfully broadcasted tx")

		return nil
	}

	telemetry.IncrCounter(1, "failure", "tx", "timeout")
	return errors.New("broadcasting tx timed out")
}

func (oc RelayerClient) BroadcastContractQuery(ctx context.Context, timeout time.Duration, queries ...SmartQuery) ([]QueryResponse, error) {
	grpcConn, err := grpc.Dial(
		oc.QueryRpc[0],
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, err
	}

	defer grpcConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	g, _ := errgroup.WithContext(ctx)

	queryClient := wasmtypes.NewQueryClient(grpcConn)

	var responses []QueryResponse
	var mut sync.Mutex
	for _, query := range queries {
		func(queryMap SmartQuery) {
			g.Go(func() error {
				queryResponse, err := queryClient.SmartContractState(ctx, &queryMap.QueryMsg)
				if err != nil {
					return err
				}

				mut.Lock()
				responses = append(responses, QueryResponse{QueryType: queryMap.QueryType, QueryResponse: *queryResponse})
				mut.Unlock()

				return nil
			})
		}(query)
	}

	err = g.Wait()
	return responses, err
}

// CreateClientContext creates an SDK client Context instance used for transaction
// generation, signing and broadcasting.
func (oc RelayerClient) CreateClientContext() (client.Context, error) {
	var keyringInput io.Reader
	if len(oc.KeyringPass) > 0 {
		keyringInput = newPassReader(oc.KeyringPass)
	} else {
		keyringInput = os.Stdin
	}

	kr, err := keyring.New("relayer", oc.KeyringBackend, oc.KeyringDir, keyringInput, oc.Encoding.Marshaler)
	if err != nil {
		return client.Context{}, err
	}

	httpClient, err := tmjsonclient.DefaultHTTPClient(oc.TMRPC[0])
	if err != nil {
		return client.Context{}, err
	}

	httpClient.Timeout = oc.RPCTimeout

	tmRPC, err := rpchttp.NewWithClient(oc.TMRPC[0], "/websocket", httpClient)
	if err != nil {
		return client.Context{}, err
	}

	keyInfo, err := kr.KeyByAddress(oc.RelayerAddr)
	if err != nil {
		return client.Context{}, err
	}

	clientCtx := client.Context{
		ChainID:           oc.ChainID,
		InterfaceRegistry: oc.Encoding.InterfaceRegistry,
		Output:            os.Stderr,
		BroadcastMode:     flags.BroadcastSync,
		TxConfig:          oc.Encoding.TxConfig,
		AccountRetriever:  authtypes.AccountRetriever{},
		Codec:             oc.Encoding.Marshaler,
		LegacyAmino:       oc.Encoding.Amino,
		Input:             os.Stdin,
		NodeURI:           oc.TMRPC[0],
		Client:            tmRPC,
		Keyring:           kr,
		FromAddress:       oc.RelayerAddr,
		FromName:          keyInfo.Name,
		From:              keyInfo.Name,
		OutputFormat:      "json",
		UseLedger:         false,
		Simulate:          false,
		GenerateOnly:      false,
		Offline:           false,
		SkipConfirm:       true,
	}

	return clientCtx, nil
}

// CreateTxFactory creates an SDK Factory instance used for transaction
// generation, signing and broadcasting.
func (oc RelayerClient) CreateTxFactory(clientCtx *client.Context) (tx.Factory, error) {
	txFactory := tx.Factory{}.
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithChainID(oc.ChainID).
		WithTxConfig(clientCtx.TxConfig).
		WithGasAdjustment(oc.GasAdjustment).
		WithGasPrices(oc.GasPrices).
		WithKeybase(clientCtx.Keyring).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithSimulateAndExecute(true)

	return txFactory, nil
}

func GetChainTimestamp(clientCtx client.Context) (time.Time, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return time.Time{}, err
	}

	status, err := node.Status(context.Background())
	if err != nil {
		return time.Time{}, err
	}

	blockTime := status.SyncInfo.LatestBlockTime
	return blockTime, nil
}

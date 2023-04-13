package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
)

type (
	// RelayerClient defines a structure that interfaces with the smart-contract-enabled chain.
	RelayerClient struct {
		logger          zerolog.Logger
		ChainID         int64
		RPC             string
		PrivKey         string
		GasPrices       string
		client          *ethclient.Client
		relayerAddress  ethtypes.Address
		contractAddress ethtypes.Address
		ChainHeight     *ChainHeight
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
	chainID int64,
	RPC string,
	contractAddr string,
	relayerAddr string,
	GasPrices string,
	privKey string,
) (RelayerClient, error) {
	relayerClient := RelayerClient{
		logger:          logger.With().Str("module", "relayer_client").Logger(),
		ChainID:         chainID,
		RPC:             RPC,
		GasPrices:       GasPrices,
		relayerAddress:  common.HexToAddress(relayerAddr),
		contractAddress: common.HexToAddress(contractAddr),
		PrivKey:         privKey,
	}

	ethClient, err := ethclient.Dial(RPC)
	if err != nil {
		return RelayerClient{}, err
	}
	blockHeight, err := ethClient.BlockNumber(ctx)
	fmt.Println("block height", blockHeight)
	if err != nil {
		return RelayerClient{}, err
	}

	block, err := ethClient.BlockByNumber(ctx, big.NewInt(int64(blockHeight)))
	if err != nil {
		return RelayerClient{}, err
	}

	blockTime := block.Time()
	chainHeight, err := NewChainHeight(
		ctx,
		RPC,
		relayerClient.logger,
		blockHeight,
		blockTime,
	)
	if err != nil {
		return RelayerClient{}, err
	}

	relayerClient.ChainHeight = chainHeight
	relayerClient.client = ethClient

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
func (oc RelayerClient) BroadcastTx(nextBlockHeight, timeoutHeight uint64, rate []PriceFeedData, deviation []PriceFeedData, medians []PriceFeedMedianData) error {
	maxBlockHeight := nextBlockHeight + timeoutHeight
	lastCheckHeight := nextBlockHeight - 1

	// re-try tx until timeout
	for lastCheckHeight < maxBlockHeight {
		latestBlockHeight, err := oc.ChainHeight.GetBlockHeight()
		if err != nil {
			return err
		}

		if latestBlockHeight <= lastCheckHeight {
			continue
		}

		// set last check height to latest block height
		lastCheckHeight = latestBlockHeight

		oracle, err := NewOracle(oc.contractAddress, oc.client)
		if err != nil {
			return err
		}

		auth, err := oc.CreateTransactor()
		if err != nil {
			return err
		}

		session := &OracleSession{
			Contract: oracle,
			CallOpts: bind.CallOpts{
				Pending: false,
			},
			TransactOpts: bind.TransactOpts{
				From:   auth.From,
				Signer: auth.Signer,
				//GasPrice: big.NewInt(160000000000),
				//Nonce:    big.NewInt(4),
			},
		}

		resp, err := session.PostPrices(rate)
		resp, err = session.PostDeviations(deviation)
		resp, err = session.PostMedians(medians)

		if resp != nil && resp.Hash().String() == "" {
			telemetry.IncrCounter(1, "failure", "tx", "code")
			oc.logger.Error().Msg(resp.Hash().String())
		}

		if err != nil {
			var (
				hash string
			)
			if resp != nil {
				hash = resp.Hash().String()
			}

			oc.logger.Debug().
				Err(err).
				Int64("max_height", int64(maxBlockHeight)).
				Int64("last_check_height", int64(lastCheckHeight)).
				Str("tx_hash", hash).
				Msg("failed to broadcast tx; retrying...")

			time.Sleep(time.Second * 1)
			continue
		}

		oc.logger.Info().
			Str("tx_hash", resp.Hash().String()).
			Uint64("nonce", resp.Nonce()).
			Msg("successfully broadcasted tx")

		return nil
	}

	telemetry.IncrCounter(1, "failure", "tx", "timeout")
	return errors.New("broadcasting tx timed out")
}

// CreateTransactor creates an SDK client Context instance used for transaction
// generation, signing and broadcasting.
func (oc RelayerClient) CreateTransactor() (*bind.TransactOpts, error) {
	sk := crypto.ToECDSAUnsafe(common.FromHex(oc.PrivKey))
	a, err := bind.NewKeyedTransactorWithChainID(sk, big.NewInt(oc.ChainID))
	return a, err
}

package client

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/ojo-network/cw-relayer/tools"
)

type (
	// RelayerClient defines a structure that interfaces with the smart-contract-enabled chain.
	RelayerClient struct {
		logger          zerolog.Logger
		ChainID         int64
		RPC             string
		PrivKey         string
		GasPriceCap     *big.Int
		GasTipCap       *big.Int
		client          *ethclient.Client
		RelayerAddress  ethtypes.Address
		contractAddress ethtypes.Address
		ChainHeight     *ChainHeight
	}
)

type QueryResponse struct {
	PriceID     uint64
	DeviationID uint64
	MedianID    uint64
}

func NewRelayerClient(
	ctx context.Context,
	logger zerolog.Logger,
	chainID int64,
	RPC string,
	contractAddr string,
	relayerAddr string,
	GasPriceCap int64,
	GasTipCap int64,
	privKey string,
) (RelayerClient, error) {
	relayerClient := RelayerClient{
		logger:          logger.With().Str("module", "relayer_client").Logger(),
		ChainID:         chainID,
		RPC:             RPC,
		GasPriceCap:     big.NewInt(GasPriceCap),
		GasTipCap:       big.NewInt(GasTipCap),
		RelayerAddress:  common.HexToAddress(relayerAddr),
		contractAddress: common.HexToAddress(contractAddr),
		PrivKey:         privKey,
	}

	ethClient, err := ethclient.Dial(RPC)
	if err != nil {
		return RelayerClient{}, err
	}

	blockHeight, err := ethClient.BlockNumber(ctx)
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

func (oc RelayerClient) BroadcastContractQuery(ctx context.Context, assetName string) (QueryResponse, error) {
	oracle, err := NewOracle(oc.contractAddress, oc.client)
	if err != nil {
		return QueryResponse{}, err
	}

	callOpts := bind.CallOpts{
		Pending: false,
	}

	g, _ := errgroup.WithContext(ctx)

	var mut sync.Mutex
	var response QueryResponse
	asset := tools.StringToByte32(assetName)

	g.Go(func() error {
		data, err := oracle.GetPriceData(&callOpts, asset)
		if err != nil {
			return err
		}

		mut.Lock()
		response.PriceID = data.Id.Uint64()
		mut.Unlock()

		return nil
	})

	g.Go(func() error {
		data, err := oracle.GetDeviationData(&callOpts, asset)
		if err != nil {
			return err
		}

		mut.Lock()
		response.DeviationID = data.Id.Uint64()
		mut.Unlock()

		return nil
	})

	g.Go(func() error {
		data, err := oracle.GetMedianData(&callOpts, asset)
		if err != nil {
			return err
		}
		mut.Lock()
		response.MedianID = data.Id.Uint64()
		mut.Unlock()

		return nil
	})

	err = g.Wait()
	return response, err
}

// BroadcastTx attempts to broadcast a signed transaction. If it fails, a few re-attempts
// will be made until the transaction succeeds or ultimately times out or fails.
func (oc RelayerClient) BroadcastTx(nextBlockHeight, timeoutHeight uint64, rate []PriceFeedData, deviation []PriceFeedData, medians []PriceFeedMedianData, disableResolve bool) error {
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

		pending, err := oc.client.PendingNonceAt(context.Background(), oc.RelayerAddress)
		if err != nil {
			return err
		}

		session := &OracleSession{
			Contract: oracle,
			CallOpts: bind.CallOpts{
				Pending: false,
			},
			TransactOpts: bind.TransactOpts{
				From:      auth.From,
				Signer:    auth.Signer,
				Nonce:     big.NewInt(int64(pending)),
				GasFeeCap: oc.GasPriceCap,
				GasTipCap: oc.GasTipCap,
			},
		}

		respRate, err := session.PostPrices(rate, disableResolve)
		if err != nil {
			return err
		}

		session.incrementNonce()
		respDeviation, err := session.PostDeviations(deviation, disableResolve)
		if err != nil {
			return err
		}

		session.incrementNonce()
		respMedian, err := session.PostMedians(medians, disableResolve)
		if err != nil {
			return err
		}

		txResps := []*types.Transaction{respRate, respDeviation, respMedian}
		for _, resp := range txResps {
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
		}

		return nil
	}

	telemetry.IncrCounter(1, "failure", "tx", "timeout")
	return errors.New("broadcasting tx timed out")
}

// CreateTransactor creates a auth signer for geth
func (oc RelayerClient) CreateTransactor() (*bind.TransactOpts, error) {
	sk := crypto.ToECDSAUnsafe(common.FromHex(oc.PrivKey))
	return bind.NewKeyedTransactorWithChainID(sk, big.NewInt(oc.ChainID))
}

func (s *OracleSession) incrementNonce() {
	nonce := s.TransactOpts.Nonce.Int64()
	s.TransactOpts.Nonce.Set(big.NewInt(nonce + 1))
}

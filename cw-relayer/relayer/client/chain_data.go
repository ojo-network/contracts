package client

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
)

// ChainHeight is used to cache the chain height of the
// current node which is being updated each time the
// node sends a new block event.
// It starts a goroutine to subscribe to blockchain new block event and update the cached height.
type ChainHeight struct {
	Logger zerolog.Logger

	mtx             sync.RWMutex
	errGetBlock     error
	lastBlockHeight uint64
	lastBlockTime   uint64
}

// NewChainHeight returns a new ChainHeight struct that
// starts a new goroutine subscribed to new block event.
func NewChainHeight(
	ctx context.Context,
	rpcURL string,
	logger zerolog.Logger,
	initialHeight uint64,
	initialBlockTime uint64,
) (*ChainHeight, error) {
	if initialHeight < 1 {
		return nil, fmt.Errorf("expected positive initial block height")
	}

	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	headerSubscription := make(chan *types.Header)
	_, err = ethClient.SubscribeNewHead(ctx, headerSubscription)

	if err != nil {
		return nil, err
	}

	chainHeight := &ChainHeight{
		Logger:          logger.With().Str("relayer_client", "chain_height").Logger(),
		errGetBlock:     nil,
		lastBlockHeight: initialHeight,
		lastBlockTime:   initialBlockTime,
	}

	go chainHeight.subscribe(ctx, headerSubscription)

	return chainHeight, nil
}

// updateBlockHeight receives the data to be updated thread safe.
func (chainHeight *ChainHeight) updateBlockHeight(blockHeight uint64, blockTime uint64, err error) {
	chainHeight.mtx.Lock()
	defer chainHeight.mtx.Unlock()

	chainHeight.lastBlockHeight = blockHeight
	chainHeight.lastBlockTime = blockTime
	chainHeight.errGetBlock = err
}

// subscribe listens to new blocks being made
// and updates the block height.
func (chainHeight *ChainHeight) subscribe(
	ctx context.Context,
	headerSubscription <-chan *types.Header,
) {
	for {
		select {
		case <-ctx.Done():
			chainHeight.Logger.Info().Msg("closing the ChainHeight subscription")
			return

		case header := <-headerSubscription:
			blockNumber := header.Number.Uint64()
			blockTime := header.Time
			chainHeight.Logger.Debug().Uint64("block number", blockNumber).Uint64("block time", blockTime).Msg("new header")

			chainHeight.updateBlockHeight(blockNumber, blockTime, nil)
		}
	}
}

// GetBlockHeight returns the last block height available.
func (chainHeight *ChainHeight) GetBlockHeight() (uint64, error) {
	chainHeight.mtx.RLock()
	defer chainHeight.mtx.RUnlock()

	return chainHeight.lastBlockHeight, chainHeight.errGetBlock
}

// GetBlockTime returns the last block time
func (chainHeight *ChainHeight) GetBlockTime() (uint64, error) {
	chainHeight.mtx.RLock()
	defer chainHeight.mtx.RUnlock()

	return chainHeight.lastBlockTime, chainHeight.errGetBlock
}

package txbundle

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"

	"github.com/ojo-network/cw-relayer/relayer/client"
)

type Txbundle struct {
	//pendingTx
	bundleSize      int64
	maxLimitPerTx   int64
	logger          zerolog.Logger
	MsgChan         chan types.Msg
	relayerClient   client.RelayerClient
	timeoutDuration time.Duration
	timeoutHeight   int64
	msgs            []types.Msg
}

func NewTxBundler(
	logger zerolog.Logger,
	bundleSize,
	maxLimitPerTx,
	timeoutHeight int64,
	timeoutDuration time.Duration,
	client client.RelayerClient,
) *Txbundle {
	tx := &Txbundle{
		bundleSize:      bundleSize,
		maxLimitPerTx:   maxLimitPerTx,
		logger:          logger.With().Str("module", "tx-bundler").Logger(),
		MsgChan:         make(chan types.Msg, bundleSize),
		timeoutDuration: timeoutDuration,
		timeoutHeight:   timeoutHeight,
		relayerClient:   client,
	}

	return tx
}

func (tx *Txbundle) Bundler(ctx context.Context) error {
	msgs := []types.Msg{}
	for {
		select {
		case <-ctx.Done():
			return nil

		default:
			for msg := range tx.MsgChan {
				if msg == nil {
					// service closed
					return nil
				}
				msgs = append(msgs, msg)

				//TODO: faster gas
				//TODO: redis store and gas bundling logic

				if err := tx.relayerClient.BroadcastTx(tx.timeoutHeight, msgs...); err != nil {
					tx.logger.Info().Err(err).Send()
				}
				msgs = []types.Msg{}
			}
		}
	}
}

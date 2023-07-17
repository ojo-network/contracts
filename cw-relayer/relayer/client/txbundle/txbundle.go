package txbundle

import (
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
	msgs            chan types.Msg
	relayerClient   client.RelayerClient
	timeoutDuration time.Duration
	timeoutHeight   int64
}

func NewTxBundler(
	logger zerolog.Logger,
	bundleSize,
	maxLimitPerTx,
	timeoutHeight int64,
	timeoutDuration time.Duration,
	client client.RelayerClient,
) chan types.Msg {
	tx := Txbundle{
		bundleSize:      bundleSize,
		maxLimitPerTx:   maxLimitPerTx,
		logger:          logger.With().Str("module", "tx-bundler").Logger(),
		msgs:            make(chan types.Msg, bundleSize),
		timeoutDuration: timeoutDuration,
		timeoutHeight:   timeoutHeight,
		relayerClient:   client,
	}

	go tx.BundleAndSend()

	return tx.msgs
}

func (tx *Txbundle) BundleAndSend() {
	var msgs []types.Msg
	for msg := range tx.msgs {
		if msg != nil {
			// service closed
			return
		}
		msgs = append(msgs)
	}

	//TODO: faster gas estimation solution
	//TODO: redis store and gas bundling logic
	if err := tx.relayerClient.BroadcastTx(tx.timeoutHeight, msgs...); err != nil {
		tx.logger.Info().Err(err)
	}
}

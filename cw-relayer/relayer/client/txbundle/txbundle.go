package txbundle

import (
	"context"
	"time"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	gastx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
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

	clientContext     sdkclient.Context
	client            tx.ServiceClient
	txFactory         gastx.Factory
	maxGasLimitperTx  uint64
	totalGasThreshold uint64
	estimateAndBundle bool
	totalTxThreshold  int
	currentThreshold  uint64
}

func NewTxBundler(
	logger zerolog.Logger,
	bundleSize,
	maxLimitPerTx,
	timeoutHeight int64,
	timeoutDuration time.Duration,
	relayerClient client.RelayerClient,
	maxGasLimitPerTx uint64,
	totalGasThreshold uint64,
	totalTxThreshold int,
	estimateAndBundle bool,
) (*Txbundle, error) {
	txbundle := &Txbundle{
		bundleSize:        bundleSize,
		maxLimitPerTx:     maxLimitPerTx,
		logger:            logger.With().Str("module", "tx-bundler").Logger(),
		MsgChan:           make(chan types.Msg, 1000000),
		timeoutDuration:   timeoutDuration,
		timeoutHeight:     timeoutHeight,
		relayerClient:     relayerClient,
		maxGasLimitperTx:  maxGasLimitPerTx,
		totalGasThreshold: totalGasThreshold,
		totalTxThreshold:  totalTxThreshold,
	}

	if estimateAndBundle {
		clientContext, err := relayerClient.CreateClientContext()
		if err != nil {
			return nil, err
		}

		clientContext.Offline = true
		clientContext.GenerateOnly = true

		txf, _ := relayerClient.CreateTxFactory(&clientContext)
		num, seq, err := txf.AccountRetriever().GetAccountNumberSequence(clientContext, clientContext.GetFromAddress())
		if err != nil {
			return nil, err
		}

		txf = txf.WithAccountNumber(num)
		txf = txf.WithSequence(seq)
		txSvcClient := tx.NewServiceClient(clientContext)

		txbundle.client = txSvcClient
		txbundle.clientContext = clientContext
		txbundle.txFactory = txf
		txbundle.estimateAndBundle = estimateAndBundle
	}

	return txbundle, nil
}

func (b *Txbundle) Bundler(ctx context.Context) error {
	ticker := time.NewTicker(b.timeoutDuration)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if len(b.msgs) > 0 {
				b.logger.Info().Msg("timeout exceeded for gas or limit threshold")
				err := b.broadcast()
				if err != nil {
					b.logger.Err(err).Send()
				}
			}
		default:
			for msg := range b.MsgChan {
				if msg == nil {
					// service closed
					return nil
				}
				if b.estimateAndBundle {
					num, seq, err := b.txFactory.AccountRetriever().GetAccountNumberSequence(b.clientContext, b.clientContext.GetFromAddress())
					if err != nil {
						return err
					}

					b.txFactory = b.txFactory.WithAccountNumber(num)
					b.txFactory = b.txFactory.WithSequence(seq)

					utx, err := b.txFactory.BuildSimTx(msg)
					if err != nil {
						return err
					}

					simRes, err := b.client.Simulate(context.Background(), &tx.SimulateRequest{TxBytes: utx})
					if err != nil {
						b.logger.Err(err).Send()
						continue
					}

					gasUsed := simRes.GetGasInfo().GasUsed
					if gasUsed > b.maxGasLimitperTx {
						// dropping tx due to max gas
						continue
					}

					// send prev msgs if the gasUsed here exceeds the total gas threshold
					if b.currentThreshold+gasUsed > b.totalGasThreshold {
						err := b.broadcast()
						if err != nil {
							b.logger.Err(err).Send()
						}

						b.msgs = []types.Msg{}
					}

					b.msgs = append(b.msgs, msg)
				} else {
					b.msgs = append(b.msgs, msg)
					if len(b.msgs) >= b.totalTxThreshold {
						err := b.broadcast()
						if err != nil {
							b.logger.Err(err).Send()
						}

						b.msgs = []types.Msg{}
					}
				}
			}
		}
	}
}

func (b *Txbundle) broadcast() error {
	b.logger.Info().Msg("broadcasting txs")
	err := b.relayerClient.BroadcastTx(b.timeoutHeight, b.msgs...)
	if err != nil {
		return err
	}

	return nil
}

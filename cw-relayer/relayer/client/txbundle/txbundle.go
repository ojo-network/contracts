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

// Txbundle is a struct that holds the necessary data and settings
// for bundling and broadcasting transactions.
type Txbundle struct {
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
	PingChan          chan types.Msg
}

// NewTxBundler initializes the Txbundle struct with the provided parameters and creates a client context,
// a transaction factory and a service client if estimateAndBundle is set to true.
func NewTxBundler(
	logger zerolog.Logger,
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
		maxLimitPerTx:     maxLimitPerTx,
		logger:            logger.With().Str("module", "tx-bundler").Logger(),
		MsgChan:           make(chan types.Msg, totalTxThreshold*2),
		PingChan:          make(chan types.Msg, 1),
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

// Bundler is a method on the Txbundle struct. It constantly listens for incoming messages
// and either directly broadcasts them or adds them to the bundle based on the estimateAndBundle setting.
// Bundles are either broadcasted when the time limit is exceeded, the total gas threshold is reached or
// when the total number of transactions in the bundle reaches the maximum limit.
func (b *Txbundle) Bundler(ctx context.Context) error {
	ticker := time.NewTicker(b.timeoutDuration)
	for {
		select {
		case <-ctx.Done():
			close(b.MsgChan)
			return nil
		case <-ticker.C:
			if len(b.msgs) > 0 {
				b.logger.Info().Msg("timeout exceeded for bundling tx, broadcasting")
				err := b.broadcast()
				if err != nil {
					b.logger.Err(err).Send()
				}
			}
		case msg, ok := <-b.PingChan:
			if !ok {
				b.logger.Warn().Msg("ping channel closed")
			}

			err := b.relayerClient.BroadcastTx(b.timeoutHeight, msg)
			if err != nil {
				b.logger.Err(err).Send()
			}

		case msg, ok := <-b.MsgChan:
			if !ok {
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
				}

				b.msgs = append(b.msgs, msg)
			} else {
				b.msgs = append(b.msgs, msg)
				if len(b.msgs) >= b.totalTxThreshold {
					err := b.broadcast()
					if err != nil {
						b.logger.Err(err).Send()
					}
				}
			}
		}
	}
}

func (b *Txbundle) broadcast() error {
	b.logger.Info().Int("total tx", len(b.msgs)).Msg("broadcasting txs")
	err := b.relayerClient.BroadcastTx(b.timeoutHeight, b.msgs...)

	b.msgs = []types.Msg{}
	if err != nil {
		return err
	}

	return nil
}

package client

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmjsonclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	wsEndpoint = "/websocket"
)

var (
	errParseEventDataNewBlockHeader = errors.New("error parsing EventDataNewBlockHeader")
	queryEventNewBlockHeader        = tmtypes.QueryForEvent(tmtypes.EventNewBlockHeader)
)

type ChainSubscribe struct {
	logger zerolog.Logger

	mtx            sync.RWMutex
	maxTickTimeout time.Duration
	rpcAddress     []string
	index          int
	rpcClient      *rpchttp.HTTP
	timeout        time.Duration
	eventChan      <-chan tmctypes.ResultEvent
	Tick           chan struct{}

	lastChainHeight    int64
	lastBlockTimestamp time.Time

	updateError error
}

// NewChainSubscription returns a new ChainSubscribe for block height update
func NewChainSubscription(
	ctx context.Context,
	logger zerolog.Logger,
	rpcAddress []string,
	timeout time.Duration,
	maxTickTimeout time.Duration,
	tickEventType string,
	skipError bool,
	maxRetries int64,
) (*ChainSubscribe, error) {
	newEvent := &ChainSubscribe{
		logger: logger.With().Str("event", tickEventType).Logger(),
		// assuming 15-second price update
		Tick:           make(chan struct{}, 4),
		timeout:        timeout,
		maxTickTimeout: maxTickTimeout,
		rpcAddress:     rpcAddress,
	}

	err := newEvent.setNewEventChan(ctx)
	if err != nil {
		if !skipError {
			return nil, err
		}

		// loop through all rpcs to connect until max retry threshold
		for i := int64(0); ; i++ {
			if i >= maxRetries {
				newEvent.logger.Err(err).Msg("error connecting to rpc")
				return nil, err
			}

			err = newEvent.switchRpc(ctx)
			if err == nil {
				break
			}
		}
	}

	go newEvent.subscribe(ctx, tickEventType)

	return newEvent, nil
}

// setNewEventChan subscribes to tendermint rpc for a specific event
func (event *ChainSubscribe) setNewEventChan(ctx context.Context) error {
	event.logger.Info().Str("new rpc", event.rpcAddress[event.index]).Msg("connecting to rpc")
	httpClient, err := tmjsonclient.DefaultHTTPClient(event.rpcAddress[event.index])
	if err != nil {
		return err
	}

	httpClient.Timeout = event.timeout
	rpcClient, err := rpchttp.NewWithClient(
		event.rpcAddress[event.index],
		wsEndpoint,
		httpClient,
	)
	if err != nil {
		return err
	}

	if !rpcClient.IsRunning() {
		if err := rpcClient.Start(); err != nil {
			return err
		}
	}

	status, err := rpcClient.Status(ctx)
	if err != nil {
		return err
	}

	// update chain height and timestamp
	event.updateChainStat(status.SyncInfo.LatestBlockHeight, status.SyncInfo.LatestBlockTime, nil)
	event.rpcClient = rpcClient

	ctx, cancel := context.WithTimeout(ctx, event.timeout)
	defer cancel()

	// tendermint overrides subscriber param
	newSubscription, err := rpcClient.Subscribe(ctx, "", queryEventNewBlockHeader.String())
	if err != nil {
		return err
	}
	event.eventChan = newSubscription

	return nil
}

// subscribe listens to new blocks being made
// and updates the chain height.
func (event *ChainSubscribe) subscribe(
	ctx context.Context,
	tickEventType string,
) {
	current := time.Now()
	for {
		select {
		case <-ctx.Done():
			err := event.rpcClient.Unsubscribe(ctx, "", queryEventNewBlockHeader.String())
			if err != nil {
				event.logger.Err(err).Msg("unsubscribing error")
			}

			event.logger.Info().Msg("closing the event subscription")
			close(event.Tick)

			return

		case resultEvent := <-event.eventChan:
			data, ok := resultEvent.Data.(tmtypes.EventDataNewBlockHeader)
			if !ok {
				event.logger.Error().Msg("no new block header")
				event.updateChainStat(event.lastChainHeight, event.lastBlockTimestamp, errParseEventDataNewBlockHeader)
				continue
			}

			event.updateChainStat(data.Header.Height, data.Header.Time, nil)
			events := data.ResultEndBlock.GetEvents()
			if len(events) > 0 {
				tick := false
				for _, event := range events {
					if event.Type == tickEventType {
						tick = true
						break
					}
				}

				if tick {
					current = time.Now()
					event.logger.Info().Msg("price update event")
					event.Tick <- struct{}{}
				}
			}

		default:
			lapsed := time.Since(current)
			if lapsed.Seconds() > event.maxTickTimeout.Seconds() {
				// reconnect to different rpc
				event.logger.Warn().Msgf("no tick since %v seconds", lapsed.Seconds())

				// is rpc client is running, unsubscribe and stop
				if event.rpcClient.IsRunning() {
					err := event.rpcClient.UnsubscribeAll(ctx, "")
					if err != nil {
						event.logger.Err(err).Msg("error unsubscribing events")
						continue
					}

					err = event.rpcClient.Stop()
					if err != nil {
						event.logger.Err(err).Msg("error stopping previous rpc client")
						continue
					}
				}

				// switching to alternative
				err := event.switchRpc(ctx)
				if err != nil {
					event.logger.Err(err).Msg("error switching to new rpc")
					continue
				}

				current = time.Now()
			}
		}
	}
}

func (event *ChainSubscribe) updateChainStat(blockHeight int64, timeStamp time.Time, err error) {
	event.mtx.Lock()
	defer event.mtx.Unlock()

	event.lastChainHeight = blockHeight
	event.lastBlockTimestamp = timeStamp
	event.updateError = err
}

// GetChainHeight returns the last chain height available.
func (event *ChainSubscribe) GetChainHeight() (int64, error) {
	event.mtx.RLock()
	defer event.mtx.RUnlock()

	return event.lastChainHeight, event.updateError
}

// GetChainTimestamp returns the last block timestamp
func (event *ChainSubscribe) GetChainTimestamp() (time.Time, error) {
	event.mtx.RLock()
	defer event.mtx.RUnlock()

	return event.lastBlockTimestamp, event.updateError
}

func (event *ChainSubscribe) switchRpc(ctx context.Context) error {
	event.index = (event.index + 1) % len(event.rpcAddress)
	err := event.setNewEventChan(ctx)

	return err
}

package client

import (
	"context"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmjsonclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/rs/zerolog"
)

const (
	wsEndpoint = "/websocket"
)

type EventSubscribe struct {
	logger         zerolog.Logger
	maxTickTimeout time.Duration
	rpcAddress     []string
	index          int
	rpcClient      *rpchttp.HTTP
	timeout        time.Duration
	eventChan      <-chan tmctypes.ResultEvent
	Tick           chan struct{}
}

func NewBlockHeightSubscription(
	ctx context.Context,
	rpcAddress []string,
	timeout time.Duration,
	maxTickTimeout time.Duration,
	tickEventType string,
	logger zerolog.Logger,
	skipError bool,
	maxRetries int64,
) (*EventSubscribe, error) {
	newEvent := &EventSubscribe{
		logger: logger.With().Str("event", tickEventType).Logger(),
		// assuming 15-second price update
		Tick:           make(chan struct{}, 100),
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

// setNewEventChan subscribes to cometbft rpc for a specific event
func (event *EventSubscribe) setNewEventChan(ctx context.Context) error {
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

	event.rpcClient = rpcClient
	eventType := tmtypes.EventNewBlockHeader
	queryType := tmtypes.QueryForEvent(eventType).String()

	ctx, cancel := context.WithTimeout(ctx, event.timeout)
	defer cancel()

	// cometbft overrides subscriber param
	newSubscription, err := rpcClient.Subscribe(ctx, "", queryType)
	if err != nil {
		return err
	}
	event.eventChan = newSubscription

	return nil
}

// subscribe listens to new blocks being made
// and updates the chain height.
func (event *EventSubscribe) subscribe(
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
				continue
			}

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
				event.logger.Info().Msgf("no tick since %v seconds", lapsed.Seconds())

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

func (event *EventSubscribe) switchRpc(ctx context.Context) error {
	event.index = (event.index + 1) % len(event.rpcAddress)
	err := event.setNewEventChan(ctx)

	return err
}

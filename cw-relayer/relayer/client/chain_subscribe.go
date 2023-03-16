package client

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	tmrpcclient "github.com/tendermint/tendermint/rpc/client"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmjsonclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	wsEndpoint = "/websocket"
)

type EventSubscribe struct {
	logger zerolog.Logger
	Tick   chan struct{}
}

func NewBlockHeightSubscription(
	ctx context.Context,
	rpcAddress string,
	timeout time.Duration,
	tickEventType string,
	logger zerolog.Logger,
) (chan struct{}, error) {
	httpClient, err := tmjsonclient.DefaultHTTPClient(rpcAddress)
	if err != nil {
		return nil, err
	}

	httpClient.Timeout = timeout
	rpcClient, err := rpchttp.NewWithClient(rpcAddress, wsEndpoint, httpClient)
	if err != nil {
		return nil, err
	}

	if !rpcClient.IsRunning() {
		if err := rpcClient.Start(); err != nil {
			return nil, err
		}
	}

	eventType := tmtypes.EventNewBlockHeader
	queryType := tmtypes.QueryForEvent(eventType).String()
	newSubscription, err := rpcClient.Subscribe(ctx, eventType, queryType)
	if err != nil {
		return nil, err
	}

	newEvent := &EventSubscribe{
		logger: logger.With().Str("chain_event_client", eventType).Logger(),
	}

	go newEvent.subscribe(ctx, rpcClient, queryType, tickEventType, newSubscription)
	newEvent.Tick = make(chan struct{})

	return newEvent.Tick, nil
}

// subscribe listens to new blocks being made
// and updates the chain height.
func (event *EventSubscribe) subscribe(
	ctx context.Context,
	eventsClient tmrpcclient.EventsClient,
	queryType string,
	tickEventType string,
	newBlockHeader <-chan tmctypes.ResultEvent,
) {
	for {
		select {
		case <-ctx.Done():
			err := eventsClient.Unsubscribe(ctx, queryType, queryEventNewBlockHeader.String())
			if err != nil {
				event.logger.Err(err).Msg("unsubscribing error")
			}

			event.logger.Info().Msg("closing the event subscription")
			close(event.Tick)

			return

		case resultEvent := <-newBlockHeader:
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
					event.logger.Info().Msg("price update event")
					event.Tick <- struct{}{}
				}
			}
		}
	}
}

package client

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	tmrpcclient "github.com/tendermint/tendermint/rpc/client"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmjsonclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	tmtypes "github.com/tendermint/tendermint/types"
)

type EventSubscribe struct {
	Logger zerolog.Logger
	Tick   chan struct{}
}

func NewEventSubscribe(
	ctx context.Context,
	rpcAddress string,
	logger zerolog.Logger,
) (*EventSubscribe, error) {
	httpClient, err := tmjsonclient.DefaultHTTPClient(rpcAddress)
	if err != nil {
		return nil, err
	}

	httpClient.Timeout = 2 * time.Minute
	rpcClient, err := rpchttp.NewWithClient(rpcAddress, "/websocket", httpClient)
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
		Logger: logger.With().Str("relayer_client", eventType).Logger(),
	}

	go newEvent.subscribe(ctx, rpcClient, queryType, newSubscription)
	newEvent.Tick = make(chan struct{})

	return newEvent, nil
}

// subscribe listens to new blocks being made
// and updates the chain height.
func (event *EventSubscribe) subscribe(
	ctx context.Context,
	eventsClient tmrpcclient.EventsClient,
	eventType string,
	newBlockHeader <-chan tmctypes.ResultEvent,
) {
	for {
		select {
		case <-ctx.Done():
			err := eventsClient.Unsubscribe(ctx, eventType, queryEventNewBlockHeader.String())
			if err != nil {
				event.Logger.Err(err)
			}
			event.Logger.Info().Msg("closing the event subscription")
			return

		case resultEvent := <-newBlockHeader:
			data, ok := resultEvent.Data.(tmtypes.EventDataNewBlockHeader)
			if !ok {
				event.Logger.Err(errors.New("no new block"))
				continue
			}

			if len(data.ResultEndBlock.GetEvents()) > 0 {
				event.Tick <- struct{}{}
			}
		}
	}
}

package client

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client/http"
	tmtypes "github.com/tendermint/tendermint/types"
	"sync"
)

const queryType = "wasm-price-feed"

type (
	ContractSubscribe struct {
		Logger          zerolog.Logger
		query           string
		contractAddress string
		client          *rpcclient.HTTP
		priceRequest    map[string][]PriceRequest
		mut             sync.Mutex
		Out             chan struct{}
	}

	PriceRequest struct {
		EventContractAddress string
		ResolveTime          string
		RequestedSymbol      string
		CallbackData         string
		RequestID            string
	}
)

func NewContractSubscribe(
	[]nodeURL string,
	contractAddress string,
	relayerAddress string,
	logger zerolog.Logger,
) (*ContractSubscribe, error) {
	client, err := rpcclient.New(nodeURL, "/websocket")
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("%s._contract_address='%s AND %s.relayer_address=%s'", queryType, contractAddress, queryType, relayerAddress)

	contractSubscribe := &ContractSubscribe{
		Logger:          logger.With().Str("relayer_client", "chain_height").Logger(),
		query:           query,
		contractAddress: contractAddress,
		client:          client,
		Out:             make(chan struct{}),
	}

	return contractSubscribe, nil
}

// Subscribe listens to new blocks being made
// and updates the chain height.
func (cs *ContractSubscribe) Subscribe(
	ctx context.Context,
) error {
	err := cs.client.Start()
	if err != nil {
		panic(err)
	}

	defer cs.client.Stop()
	out, err := cs.client.Subscribe(context.Background(), "0", cs.query)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-ctx.Done():
			err := cs.client.Unsubscribe(ctx, "0", cs.query)
			if err != nil {
				cs.Logger.Err(err)
			}
			cs.Logger.Info().Msg("closing the ChainHeight subscription")
			return nil

		case result := <-out:
			switch event := result.Data.(type) {
			case tmtypes.EventDataTx:
				for _, e := range event.Result.Events {
					if e.Type == queryType {
						priceRequests := parseEvents(e.Attributes)
						cs.setPriceRequest(priceRequests)
					}
				}

			default:
				cs.Logger.Error().Err(fmt.Errorf("Unknown event type: %T\n", event))
			}
		}
	}
}

func (cs *ContractSubscribe) setPriceRequest(priceRequests []PriceRequest) {
	cs.mut.Lock()
	cs.mut.Unlock()
	prices := make(map[string][]PriceRequest)
	for _, request := range priceRequests {
		prices[request.RequestedSymbol] = append(prices[request.RequestedSymbol], request)
	}

	cs.priceRequest = prices
}

func (cs *ContractSubscribe) GetPriceRequest() map[string][]PriceRequest {
	cs.mut.Lock()
	cs.mut.Unlock()
	priceRequests := cs.priceRequest

	cs.priceRequest = make(map[string][]PriceRequest)
	return priceRequests
}

func parseEvents(attrs []abcitypes.EventAttribute) (priceRequest []PriceRequest) {
	for i := 0; i < len(attrs); i += 6 {
		req := PriceRequest{}
		for _, attr := range attrs[i : i+7] {
			val := string(attr.Value)
			switch string(attr.Key) {
			case "_contract_address":
				continue
			case "callback_data":
				req.CallbackData = val
			case "event_contract_address":
				req.EventContractAddress = val
			case "resolve_time":
				req.ResolveTime = val
			case "request_id":
				req.RequestID = val
			case "symbol":
				req.RequestedSymbol = val
			default:
				panic("should never happen")
			}
		}
		priceRequest = append(priceRequest, req)
	}

	return
}

func aggeregatePriceRequests(priceRequests []PriceRequest) map[string][]PriceRequest {
	prices := make(map[string][]PriceRequest)
	for _, request := range priceRequests {
		prices[request.RequestedSymbol] = append(prices[request.RequestedSymbol], request)
	}

	//TODO: do some price aggregation and time

	return prices
}

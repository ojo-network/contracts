package client

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client/http"
	tmtypes "github.com/tendermint/tendermint/types"
)

const queryType = "wasm-price-feed"

type EventType int

const (
	RequestRate EventType = iota
	RequestMedian
	RequestDeviation
)

type (
	ContractSubscribe struct {
		nodeURL         []string
		Logger          zerolog.Logger
		query           string
		contractAddress string
		client          *rpcclient.HTTP
		priceRequest    map[string][]PriceRequest
		mut             sync.Mutex
		Out             chan struct{}
	}

	PriceRequest struct {
		Event                EventType
		EventContractAddress string
		ResolveTime          string
		RequestedSymbol      string
		CallbackData         string
		CallbackSig          string
		RequestID            string
	}
)

func NewContractSubscribe(
	nodeURL []string,
	contractAddress string,
	relayerAddress string,
	logger zerolog.Logger,
) (*ContractSubscribe, error) {
	client, err := rpcclient.New(nodeURL[0], "/websocket")
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("%s._contract_address='%s' AND %s.relayer_address='%s'", queryType, contractAddress, queryType, relayerAddress)
	contractSubscribe := &ContractSubscribe{
		Logger:          logger.With().Str("relayer_client", "chain_height").Logger(),
		query:           query,
		contractAddress: contractAddress,
		priceRequest:    make(map[string][]PriceRequest),
		client:          client,
		nodeURL:         nodeURL,
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
			return cs.client.Stop()

		case result := <-out:
			switch event := result.Data.(type) {
			case tmtypes.EventDataTx:
				for _, e := range event.Result.Events {
					if e.Type == queryType {
						priceRequests, err := parseEvents(e.Attributes)
						if err != nil {
							cs.Logger.Error().Err(err).Send()
						}

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
	defer cs.mut.Unlock()
	prices := make(map[string][]PriceRequest)
	for _, request := range priceRequests {
		if _, ok := cs.priceRequest[request.RequestedSymbol]; !ok {
			cs.priceRequest[request.RequestedSymbol] = []PriceRequest{request}
		}

		cs.priceRequest[request.RequestedSymbol] = append(prices[request.RequestedSymbol], request)
	}
}

func (cs *ContractSubscribe) GetPriceRequest() map[string][]PriceRequest {
	cs.mut.Lock()
	defer cs.mut.Unlock()
	priceRequests := cs.priceRequest

	cs.priceRequest = make(map[string][]PriceRequest)
	return priceRequests
}

func parseEvents(attrs []abcitypes.EventAttribute) (priceRequest []PriceRequest, err error) {
	for i := 0; i < len(attrs); i += 9 {
		req := PriceRequest{}
		for _, attr := range attrs[i : i+9] {
			val := string(attr.Value)
			switch string(attr.Key) {
			case "_contract_address":
				continue
			case "request_type":
				switch val {
				case "request_rate":
					req.Event = RequestRate
				case "request_median":
					req.Event = RequestMedian
				case "request_deviation":
					req.Event = RequestDeviation
				}
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
			case "callback_signature":
				req.CallbackSig = val
			case "relayer_address":
				continue
			default:
				err = fmt.Errorf("unknown attribute: %s", attr.Key)
			}
		}

		priceRequest = append(priceRequest, req)
	}

	return
}

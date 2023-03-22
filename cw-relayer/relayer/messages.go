package relayer

import (
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/ojo-network/cw-relayer/relayer/client"
)

type MsgType int

const (
	RelayRate MsgType = iota + 1
	RelayHistoricalMedian
	RelayHistoricalDeviation
	QueryRateMsg
	QueryMedianRateMsg
	QueryDeviationRateMsg
)

func (m MsgType) String() string {
	return [...]string{"relay", "relay_historical_median", "relay_historical_deviation"}[m-1]
}

type (
	MsgRelay struct {
		Relay Msg `json:"relay"`
	}

	MsgForceRelay struct {
		Relay Msg `json:"force_relay"`
	}

	MsgRelayHistoricalMedian struct {
		Relay Msg `json:"relay_historical_median"`
	}

	MsgForceRelayHistoricalMedian struct {
		Relay Msg `json:"force_relay_historical_median"`
	}

	MsgRelayHistoricalDeviation struct {
		Relay Msg `json:"relay_historical_deviation"`
	}

	MsgForceRelayHistoricalDeviation struct {
		Relay Msg `json:"force_relay_historical_deviation"`
	}

	Msg struct {
		SymbolRates [][2]interface{} `json:"symbol_rates,omitempty"`
		ResolveTime int64            `json:"resolve_time,string"`
		RequestID   uint64           `json:"request_id,string"`
	}

	// for restart queries
	rateMsg struct {
		Ref symbol `json:"get_ref"`
	}

	medianRateMsg struct {
		Ref symbol `json:"get_median_ref"`
	}

	deviationRateMsg struct {
		Ref symbol `json:"get_deviation_ref"`
	}

	symbol struct {
		Symbol string `json:"symbol"`
	}
)

func restartQuery(contractAddress, Denom string) ([]client.SmartQuery, error) {
	rateData, err := json.Marshal(rateMsg{Ref: symbol{Symbol: Denom}})
	if err != nil {
		return nil, err
	}

	medianData, err := json.Marshal(medianRateMsg{Ref: symbol{Denom}})
	if err != nil {
		return nil, err
	}

	deviationData, err := json.Marshal(deviationRateMsg{Ref: symbol{Denom}})
	if err != nil {
		return nil, err
	}

	return []client.SmartQuery{
		{
			QueryType: int(QueryRateMsg),
			QueryMsg: wasmtypes.QuerySmartContractStateRequest{
				Address:   contractAddress,
				QueryData: rateData,
			},
		},
		{
			QueryType: int(QueryMedianRateMsg),
			QueryMsg: wasmtypes.QuerySmartContractStateRequest{
				Address:   contractAddress,
				QueryData: medianData,
			},
		},
		{
			QueryType: int(QueryDeviationRateMsg),
			QueryMsg: wasmtypes.QuerySmartContractStateRequest{
				Address:   contractAddress,
				QueryData: deviationData,
			},
		},
	}, nil
}

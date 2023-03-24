package relayer

import (
	"encoding/json"

	"github.com/ojo-network/cw-relayer/relayer/client"
	scrttypes "github.com/ojo-network/cw-relayer/relayer/dep/utils/types"
)

type (
	MsgRelay struct {
		Relay Msg `json:"relay"`
	}

	MsgForceRelay struct {
		Relay Msg `json:"force_relay"`
	}

	Msg struct {
		SymbolRates [][2]string `json:"symbol_rates,omitempty"`
		ResolveTime int64       `json:"resolve_time,string"`
		RequestID   uint64      `json:"request_id,string"`
	}

	// for restart queries
	rateMsg struct {
		Ref symbol `json:"get_ref"`
	}

	symbol struct {
		Symbol string `json:"symbol"`
	}
)

func restartQuery(contractAddress, Denom string) (client.SmartQuery, error) {
	rateData, err := json.Marshal(rateMsg{Ref: symbol{Symbol: Denom}})
	if err != nil {
		return client.SmartQuery{}, err
	}

	return client.SmartQuery{
		QueryMsg: scrttypes.QuerySecretContractRequest{
			ContractAddress: contractAddress,
			Query:           rateData,
		},
	}, err
}

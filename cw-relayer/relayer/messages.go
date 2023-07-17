package relayer

import (
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

type (
	CallbackData struct {
		RequestID    string `json:"request_id"`
		Symbol       string `json:"symbol"`
		SymbolRate   string `json:"symbol_rate"`
		LastUpdated  string `json:"last_updated"`
		CallbackData []byte `json:"callback_data"`
	}

	CallbackDataMedian struct {
		RequestID    string   `json:"request_id"`
		Symbol       string   `json:"symbol"`
		SymbolRates  []string `json:"symbol_rates"`
		LastUpdated  string   `json:"last_updated"`
		CallbackData []byte   `json:"callback_data"`
	}

	CallbackRate struct {
		Req CallbackData `json:"callback_rate_data"`
	}

	CallbackMedian struct {
		Req CallbackDataMedian `json:"callback_rate_median"`
	}

	CallbackDeviation struct {
		Req CallbackData `json:"callback_rate_deviation"`
	}

	Execute struct {
		Callback interface{} `json:"execute"`
	}

	Ping struct {
		Ping struct{} `json:"relayer_ping"`
	}
)

func genMsg(relayerAddress, contractAddress string, msg any) (*wasmtypes.MsgExecuteContract, error) {
	msgData, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return &wasmtypes.MsgExecuteContract{
		Sender:   relayerAddress,
		Contract: contractAddress,
		Msg:      msgData,
		Funds:    nil,
	}, nil
}

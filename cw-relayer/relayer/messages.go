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

	Callback struct {
		Req CallbackData `json:"callback"`
	}

	Execute struct {
		Callback Callback `json:"execute"`
	}

	Ping struct {
		Ping struct{} `json:"ping"`
	}

	Query struct {
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

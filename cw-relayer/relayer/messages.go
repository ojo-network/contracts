package relayer

import (
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

type (
	// CallbackData for price relay support
	CallbackData struct {
		RequestID    string `json:"request_id"`
		Symbol       string `json:"symbol"`
		SymbolRate   string `json:"symbol_rate"`
		LastUpdated  string `json:"last_updated"`
		CallbackData []byte `json:"callback_data"`
	}

	// CallbackDataHistorical for Median and Deviation relay support
	CallbackDataHistorical struct {
		RequestID    string   `json:"request_id"`
		Symbol       string   `json:"symbol"`
		SymbolRates  []string `json:"symbol_rates"`
		LastUpdated  string   `json:"last_updated"`
		CallbackData []byte   `json:"callback_data"`
	}

	// Ping msg type relayer uptime ping
	Ping struct {
		Ping struct{} `json:"relayer_ping"`
	}
)

// genMsg generates a wasmtype MsgExecuteContract msg with a particular callback signature according to a contract request
func genMsg(relayerAddress, contractAddress, callbackSig string, msg any) (*wasmtypes.MsgExecuteContract, error) {
	execute := make(map[string]interface{})
	execute[callbackSig] = msg

	msgData, err := json.Marshal(execute)
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

// genPingMsg generates a wasmtype MsgExecuteContract msg for relayer ping
func genPingMsg(relayerAddress, contractAddress string) (*wasmtypes.MsgExecuteContract, error) {
	ping := Ping{Ping: struct{}{}}
	msgData, err := json.Marshal(ping)
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

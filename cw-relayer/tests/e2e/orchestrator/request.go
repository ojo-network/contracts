package orchestrator

import (
	"encoding/json"
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

type (
	RequestType int
	GetPrice    struct {
		SymbolRequested Symbol `json:"get_price"`
	}

	Symbol struct {
		Symbol string `json:"symbol"`
	}
)

const (
	Price RequestType = iota
	Median
	Deviation
)

func (r RequestType) String() string {
	switch r {
	case Price:
		return "request_rate"
	case Median:
		return "request_median"
	case Deviation:
		return "request_deviation"
	}

	return ""
}

func (o *Orchestrator) RequestMsg(request RequestType, denom string) error {
	addMsg := fmt.Sprintf("{\"%s\":{\"symbol\":\"%s\",\"callback_data\":\"test\"}}", request.String(), denom)
	msg := []string{
		"wasmd", "tx", "wasm", "execute", o.QueryContractAddress, addMsg,
		"--from=user", "-b=block", "--gas-prices=0.25stake", "--keyring-backend=test", "--gas=auto", "--gas-adjustment=1.3", "-y",
		fmt.Sprintf("--chain-id=%s", o.WasmChain.chainId),
		"--home=/.wasmd",
	}

	return o.execWasmCmd(msg)
}

func (o *Orchestrator) GenerateQuery(request RequestType, symbol string) *wasmtypes.QuerySmartContractStateRequest {
	data := Symbol{Symbol: symbol}
	msg := make(map[string]interface{})
	switch request {
	case Price:
		msg["get_price"] = data
	case Median:
		msg["get_median"] = data
	case Deviation:
		msg["get_deviation"] = data
	}

	jsonMsg, _ := json.Marshal(msg)
	return &wasmtypes.QuerySmartContractStateRequest{
		Address:   o.QueryContractAddress,
		QueryData: jsonMsg,
	}
}

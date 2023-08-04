package orchestrator

import (
	"encoding/json"
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

const (
	Price RequestType = iota
	Median
	Deviation
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

func (r RequestType) PriceQueryString() string {
	switch r {
	case Price:
		return "get_price"
	case Median:
		return "get_median"
	case Deviation:
		return "get_deviation"
	}

	return ""
}

func (r RequestType) RequestIDQueryString() string {
	switch r {
	case Price:
		return "get_rate_request_id"
	case Median:
		return "get_median_request_id"
	case Deviation:
		return "get_deviation_request_id"
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

func (o *Orchestrator) GeneratePriceQuery(request RequestType, symbol string) *wasmtypes.QuerySmartContractStateRequest {
	data := Symbol{Symbol: symbol}
	msg := map[string]interface{}{
		request.PriceQueryString(): data,
	}

	jsonMsg, _ := json.Marshal(msg)
	return &wasmtypes.QuerySmartContractStateRequest{
		Address:   o.QueryContractAddress,
		QueryData: jsonMsg,
	}
}

func (o *Orchestrator) GenerateRequestIDQuery(request RequestType, symbol string) *wasmtypes.QuerySmartContractStateRequest {
	data := Symbol{Symbol: symbol}
	msg := map[string]interface{}{
		request.RequestIDQueryString(): data,
	}

	jsonMsg, _ := json.Marshal(msg)
	return &wasmtypes.QuerySmartContractStateRequest{
		Address:   o.QueryContractAddress,
		QueryData: jsonMsg,
	}
}

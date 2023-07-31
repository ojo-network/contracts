package orchestrator

import (
	"fmt"
)

type RequestType int

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

func (o *Orchestrator) RequestMsg(request RequestType, oracleAddress, denom string) error {
	addMsg := fmt.Sprintf("{\"%s\":{\"symbol\":\"%s\",\"callback_data\":\"test\"}}", request.String(), denom)
	msg := []string{
		"wasmd", "tx", "wasm", "execute", oracleAddress, addMsg,
		"--from=user", "-b=block", "--gas-prices=0.25stake", "--keyring-backend=test", "--gas=auto", "--gas-adjustment=1.3", "-y",
		fmt.Sprintf("--chain-id=%s", o.WasmChain.chainId),
		"--home=/.wasmd",
	}

	return o.execWasmCmd(msg)
}

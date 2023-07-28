package relayer

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/types"

	"github.com/ojo-network/cw-relayer/relayer/client"
	scrttypes "github.com/ojo-network/cw-relayer/relayer/dep/utils/types"
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

func genRestartQueries(contractAddress, Denom string) ([]client.SmartQuery, error) {
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
			QueryMsg: scrttypes.QuerySecretContractRequest{
				ContractAddress: contractAddress,
				Query:           rateData,
			},
		},
		{
			QueryType: int(QueryMedianRateMsg),
			QueryMsg: scrttypes.QuerySecretContractRequest{
				ContractAddress: contractAddress,
				Query:           medianData,
			},
		},
		{
			QueryType: int(QueryDeviationRateMsg),
			QueryMsg: scrttypes.QuerySecretContractRequest{
				ContractAddress: contractAddress,
				Query:           deviationData,
			},
		},
	}, nil
}

func generateContractRelayMsg(forceRelay bool, msgType MsgType, requestID uint64, resolveTime int64, rates types.DecCoins) (msgData []byte, err error) {
	msg := Msg{
		SymbolRates: nil,
		ResolveTime: resolveTime,
		RequestID:   requestID,
	}

	if msgType == RelayRate {
		for _, rate := range rates {
			msg.SymbolRates = append(msg.SymbolRates, [2]interface{}{rate.Denom, rate.Amount.Mul(RateFactor).TruncateInt().String()})
		}

		if forceRelay {
			msgData, err = json.Marshal(MsgForceRelay{Relay: msg})
		} else {
			msgData, err = json.Marshal(MsgRelay{Relay: msg})
		}

		return
	} else {
		symbolRates := map[string][]string{}
		for _, rate := range rates {
			symbolRates[rate.Denom] = append(symbolRates[rate.Denom], rate.Amount.Mul(RateFactor).TruncateInt().String())
		}

		for denom, medians := range symbolRates {
			msg.SymbolRates = append(msg.SymbolRates, [2]interface{}{denom, medians})
		}

		switch msgType {
		case RelayHistoricalMedian:
			if forceRelay {
				msgData, err = json.Marshal(MsgForceRelayHistoricalMedian{Relay: msg})
			} else {
				msgData, err = json.Marshal(MsgRelayHistoricalMedian{Relay: msg})
			}

		case RelayHistoricalDeviation:
			if forceRelay {
				msgData, err = json.Marshal(MsgForceRelayHistoricalDeviation{Relay: msg})
			} else {
				msgData, err = json.Marshal(MsgRelayHistoricalDeviation{Relay: msg})
			}
		}
	}

	return
}

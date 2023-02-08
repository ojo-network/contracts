package relayer

type MsgType int

const (
	RelayRate MsgType = iota + 1
	RelayHistoricalMedian
	RelayHistoricalDeviation
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
)

package relayer

type (
	MsgRelay struct {
		Relay Msg `json:"relay"`
	}

	MsgForceRelay struct {
		Relay Msg `json:"force_relay"`
	}

	Msg struct {
		SymbolRates [][2]string `json:"symbol_rates,omitempty"`
		ResolveTime int64       `json:"resolve_time,omitempty,string"`
		RequestID   uint64      `json:"request_id,string"`
	}
)

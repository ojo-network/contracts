package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ojo-network/cw-relayer/config"
)

func TestValidate(t *testing.T) {
	validConfig := func() config.Config {
		return config.Config{
			Account: config.Account{
				AccPrefix: "address",
				ChainID:   "test-chain",
				Address:   "cosmos1r20p6ye8m4m3zxs90mkwtv30rrlmcu9npytuah",
			},
			Keyring: config.Keyring{
				Backend: "os",
				Dir:     "~/.wasmd/",
			},
			TargetRPC: config.RPC{
				TMRPCEndpoint: []string{"http://localhost:26657"},
				RPCTimeout:    "5s",
				QueryEndpoint: []string{"http://localhost:1317"},
			},
			DataRPC: config.DataRpc{
				QueryRPCS: []string{"http://localhost:26657"},
				EventRPCS: []string{"http://localhost:1317"},
			},
			Timeout: config.Timeout{
				EventTimeout:    "30s",
				MaxTickTimeout:  "60s",
				QueryTimeout:    "15s",
				TimeoutHeight:   100,
				ProviderTimeout: "5s",
			},
			Gas: config.Gas{
				GasAdjustment: 1.1,
				GasPrices:     "0.025ustake",
				GasLimitPerTx: 200000,
			},
			TxConfig: config.TxConfig{
				BundleSize:        10,
				MaxGasLimitPerTx:  500000,
				TotalGasThreshold: 2000000,
				TotalTxThreshold:  100,
				EstimateAndBundle: true,
				MaxTimeout:        "10s",
			},
			MaxRetries:         3,
			PingDuration:       "5s",
			TickDuration:       "1s",
			NumBundle:          10,
			ContractAddress:    "wasm14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0phg4d",
			IgnoreMedianErrors: true,
			TickEventType:      "NewBlock",
			MaxGasUnits:        1000000,
			BlockHeightConfig: config.BlockHeightConfig{
				SkipError: true,
			},
		}
	}

	testCases := []struct {
		name      string
		cfg       config.Config
		expectErr bool
	}{
		{
			"valid config",
			validConfig(),
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.cfg.Validate() != nil, tc.expectErr)
		})
	}
}

func TestParseConfig_Valid(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "cw-relayer*.toml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := []byte(`
max_retries = 3
ping_duration = "5s"
tick_duration = "1s"
num_bundle = 10
contract_address ="wasm14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0phg4d"
ignore_median_errors = true
event_type = "NewBlock"
max_gas_units = 1000000

[account]
acc_prefix = "wasm"
chain_id = "wasmhub-4"
address = "wasm1r20p6ye8m4m3zxs90mkwtv30rrlmcu9npytuah"

[keyring]
backend = "os"
dir = "~/.wasm-cli/"

[target_rpc]
tmrpc_endpoint = ["http://localhost:26657"]
rpc_timeout = "5s"
query_endpoint = ["http://localhost:1317"]

[data_rpc]
query_rpcs = ["http://localhost:26657"]
event_rpcs = ["http://localhost:1317"]

[timeout]
event_timeout = "30s"
max_tick_timeout = "60s"
query_timeout = "15s"
timeout_height = 100
provider_timeout = "5s"

[gas]
gas_adjustment = 1.1
gas_prices = "0.025ustake"
gas_per_tx = 200000

[tx_config]
bundle_size = 10
max_gas_per_tx = 500000
total_gas_threshold = 2000000
total_tx_threshold = 100
estimate_and_bundle = true
max_timeout = "10s"

[block_height_config]
skip_error = true
`)
	_, err = tmpFile.Write(content)
	require.NoError(t, err)

	_, err = config.ParseConfig(tmpFile.Name())
	require.NoError(t, err)
}

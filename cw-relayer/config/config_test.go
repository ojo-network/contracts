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
				Address: "fromaddr",
				ChainID: "chain-id",
			},
			Keyring: config.Keyring{
				Backend: "test",
				Dir:     "/Users/username/.wasm",
			},
			RPC: config.RPC{
				TMRPCEndpoint: "http://localhost:26657",
				RPCTimeout:    "100ms",
			},
			GasAdjustment:   1.5,
			ContractAddress: "wasm14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0phg4d",
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
			true,
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
gas_adjustment = 1.5
contract_address = "wasm14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0phg4d"
gas_prices = "0.00025stake"
query_rpcs = ["http://localhost:26657"]
event_rpcs = ["http://localhost:26657"]

[account]
address = "wasm15nejfgcaanqpw25ru4arvfd0fwy6j8clccvwx4"
chain_id = "wasm-local-testnet"
acc_prefix = "wasm"

[keyring]
backend = "test"
dir = "/Users/username/.wasm"

[rpc]
tmrpc_endpoint = "http://localhost:26657"
query_endpoint = "localhost:9090"
rpc_timeout = "100ms"
`)
	_, err = tmpFile.Write(content)
	require.NoError(t, err)

	_, err = config.ParseConfig(tmpFile.Name())
	require.NoError(t, err)
}

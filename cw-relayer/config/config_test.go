package config_test

import (
	"io/ioutil"
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
				Dir:     "/Users/username/.ojo",
			},
			RPC: config.RPC{
				TMRPCEndpoint: "http://localhost:26657",
				GRPCEndpoint:  "localhost:9090",
				RPCTimeout:    "100ms",
			},
			GasAdjustment: 1.5,
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
	tmpFile, err := ioutil.TempFile("", "cw-relayer*.toml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := []byte(`
gas_adjustment = 1.5

[account]
address = "wasm15nejfgcaanqpw25ru4arvfd0fwy6j8clccvwx4"
chain_id = "wasm-local-testnet"

[keyring]
backend = "test"
dir = "/Users/username/.wasm"

[rpc]
tmrpc_endpoint = "http://localhost:26657"
grpc_endpoint = "localhost:9090"
rpc_timeout = "100ms"
`)
	_, err = tmpFile.Write(content)
	require.NoError(t, err)

	_, err = config.ParseConfig(tmpFile.Name())
	require.NoError(t, err)
}

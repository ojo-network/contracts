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
				ChainID: 1,
			},
			Keyring: config.Keyring{
				PrivKey: "",
			},
			RPC: config.RPC{
				WssEndpoint: "http://localhost:26657",
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
contract_address = "0x"
gas_prices = "0.00025stake"
query_rpcs = ["http://localhost:26657"]
event_rpcs = ["http://localhost:26657"]

[account]
address = "0x"
chain_id = "1"

[keyring]
priv_key="privkey"

[rpc]
wss_endpoint = "http://localhost:26657"
`)
	_, err = tmpFile.Write(content)
	require.NoError(t, err)

	_, err = config.ParseConfig(tmpFile.Name())
	require.NoError(t, err)
}

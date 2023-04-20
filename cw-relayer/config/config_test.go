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
			GasPriceCap:     1,
			GasTipCap:       1,
			ContractAddress: "0x",
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
gas_price_cap = 1
gas_tip_cap = 1
contract_address = "0x"
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

package relayer

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"

	"github.com/ojo-network/cw-relayer/relayer/client"
)

type RelayerTestSuite struct {
	suite.Suite
	relayer *Relayer
}

func (ots *RelayerTestSuite) SetupSuite() {
	ots.relayer = New(zerolog.Nop(), client.RelayerClient{}, "", 100, 5, "")
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RelayerTestSuite))
}

func (ots *RelayerTestSuite) TestStop() {
	ots.Eventually(
		func() bool {
			ots.relayer.Stop()
			return true
		},
		5*time.Second,
		time.Second,
	)
}

func (ots *RelayerTestSuite) Test_generateContractMsg() {
	exchangeRates := types.DecCoins{
		types.NewDecCoinFromDec("atom", types.MustNewDecFromStr("1.23456789")),
		types.NewDecCoinFromDec("umee", types.MustNewDecFromStr("1.23456789")),
		types.NewDecCoinFromDec("juno", types.MustNewDecFromStr("1.23456789")),
	}

	ots.Run("Relay msg", func() {
		msg, err := generateContractRelayMsg(false, 1, 1, exchangeRates)
		ots.Require().NoError(err)

		expectedRes := "{\"relay\":{\"symbol_rates\":[[\"atom\",\"1234567890000000000\"],[\"umee\",\"1234567890000000000\"],[\"juno\",\"1234567890000000000\"]],\"resolve_time\":\"1\",\"request_id\":\"1\"}}"
		msgStr := string(msg)

		ots.Require().Equal(expectedRes, msgStr)
	})

	ots.Run("Force Relay msg", func() {
		msg, err := generateContractRelayMsg(true, 1, 1, exchangeRates)
		ots.Require().NoError(err)

		expectedRes := "{\"force_relay\":{\"symbol_rates\":[[\"atom\",\"1234567890000000000\"],[\"umee\",\"1234567890000000000\"],[\"juno\",\"1234567890000000000\"]],\"resolve_time\":\"1\",\"request_id\":\"1\"}}"
		msgStr := string(msg)

		ots.Require().Equal(expectedRes, msgStr)
	})
}

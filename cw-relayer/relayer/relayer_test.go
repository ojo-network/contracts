package relayer

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type RelayerTestSuite struct {
	suite.Suite
}

func (rts *RelayerTestSuite) SetupSuite() {}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RelayerTestSuite))
}

func (rts *RelayerTestSuite) TestStop() {}

func (rts *RelayerTestSuite) Test_generateContractMsg() {
	exchangeRates := types.DecCoins{
		types.NewDecCoinFromDec("atom", types.MustNewDecFromStr("1.23456789")),
		types.NewDecCoinFromDec("umee", types.MustNewDecFromStr("1.23456789")),
		types.NewDecCoinFromDec("juno", types.MustNewDecFromStr("1.23456789")),
	}

	rts.Run("Relay msg", func() {
		msg, err := generateContractRelayMsg(false, 1, 1, exchangeRates)
		rts.Require().NoError(err)

		// price * 10**9 (USD factor in contract)
		expectedRes := "{\"relay\":{\"symbol_rates\":[[\"atom\",\"1234567890\"],[\"umee\",\"1234567890\"],[\"juno\",\"1234567890\"]],\"resolve_time\":\"1\",\"request_id\":\"1\"}}"
		msgStr := string(msg)

		rts.Require().Equal(expectedRes, msgStr)
	})

	rts.Run("Force Relay msg", func() {
		msg, err := generateContractRelayMsg(true, 1, 1, exchangeRates)
		rts.Require().NoError(err)

		expectedRes := "{\"force_relay\":{\"symbol_rates\":[[\"atom\",\"1234567890\"],[\"umee\",\"1234567890\"],[\"juno\",\"1234567890\"]],\"resolve_time\":\"1\",\"request_id\":\"1\"}}"
		msgStr := string(msg)

		rts.Require().Equal(expectedRes, msgStr)
	})
}

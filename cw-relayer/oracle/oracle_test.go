package oracle

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"

	"github.com/ojo-network/cw-relayer/oracle/client"
)

type OracleTestSuite struct {
	suite.Suite
	oracle *Oracle
}

func (ots *OracleTestSuite) SetupSuite() {
	ots.oracle = New(zerolog.Nop(), client.OracleClient{}, "", 100, "")

}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OracleTestSuite))
}

func (ots *OracleTestSuite) TestStop() {
	ots.Eventually(
		func() bool {
			ots.oracle.Stop()
			return true
		},
		5*time.Second,
		time.Second,
	)
}

func (ots *OracleTestSuite) Test_generateContractMsg() {
	exchangeRates := types.DecCoins{
		types.NewDecCoin("atom", types.NewInt(1)),
		types.NewDecCoin("umee", types.NewInt(2)),
		types.NewDecCoin("juno", types.NewInt(3)),
	}

	ots.Run("Relay msg", func() {
		msg, err := generateContractRelayMsg(false, 1, 1, exchangeRates)
		ots.Require().NoError(err)

		expectedRes := "{\"relay\":{\"symbol_rates\":[[\"atom\",\"1.000000000000000000\"],[\"umee\",\"2.000000000000000000\"],[\"juno\",\"3.000000000000000000\"]],\"resolve_time\":1,\"request_id\":1}}"
		msgStr := string(msg)

		ots.Require().Equal(expectedRes, msgStr)
	})

	ots.Run("Force Relay msg", func() {
		msg, err := generateContractRelayMsg(true, 1, 1, exchangeRates)
		ots.Require().NoError(err)

		expectedRes := "{\"force_relay\":{\"symbol_rates\":[[\"atom\",\"1.000000000000000000\"],[\"umee\",\"2.000000000000000000\"],[\"juno\",\"3.000000000000000000\"]],\"resolve_time\":1,\"request_id\":1}}"
		msgStr := string(msg)

		ots.Require().Equal(expectedRes, msgStr)
	})
}

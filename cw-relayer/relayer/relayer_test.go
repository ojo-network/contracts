package relayer

import (
	"encoding/json"
	"fmt"
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

func (rts *RelayerTestSuite) SetupSuite() {
	rts.relayer = New(zerolog.Nop(), client.RelayerClient{}, "", 100, 5, "", 0)
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RelayerTestSuite))
}

func (rts *RelayerTestSuite) TestStop() {
	rts.Eventually(
		func() bool {
			rts.relayer.Stop()
			return true
		},
		5*time.Second,
		time.Second,
	)
}

func (rts *RelayerTestSuite) Test_generateRelayMsg() {
	exchangeRates := types.DecCoins{
		types.NewDecCoinFromDec("atom", types.MustNewDecFromStr("1.23456789")),
		types.NewDecCoinFromDec("umee", types.MustNewDecFromStr("1.23456789")),
		types.NewDecCoinFromDec("juno", types.MustNewDecFromStr("1.23456789")),
	}

	testCases := []struct {
		tc         string
		forceRelay bool
		msgType    MsgType
	}{
		{
			tc:         "Relay msg",
			forceRelay: false,
			msgType:    RelayRate,
		},
		{
			tc:         "Force Relay msg",
			forceRelay: true,
			msgType:    RelayRate,
		},
		{
			tc:         "Relay median",
			forceRelay: false,
			msgType:    RelayHistoricalMedian,
		},
		{
			tc:         "Relay deviations",
			forceRelay: false,
			msgType:    RelayHistoricalDeviation,
		},
	}

	for _, tc := range testCases {
		rts.Run(tc.tc, func() {
			msg, err := genRateMsgData(tc.forceRelay, tc.msgType, 0, 0, exchangeRates)
			rts.Require().NoError(err)

			var expectedMsg map[string]Msg
			err = json.Unmarshal(msg, &expectedMsg)
			rts.Require().NoError(err)

			var msgKey string
			if tc.forceRelay {
				msgKey = fmt.Sprintf("force_%s", tc.msgType.String())
			} else {
				msgKey = tc.msgType.String()
			}

			rates := expectedMsg[msgKey].SymbolRates
			rts.Require().NotZero(len(rates))

			for i, rate := range rates {
				rts.Require().Equal(rate[0], exchangeRates[i].Denom)
				rts.Require().Equal(rate[1], exchangeRates[i].Amount.Mul(RateFactor).TruncateInt().String())
			}
		})
	}
}

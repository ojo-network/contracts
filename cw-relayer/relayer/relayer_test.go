package relayer

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
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
	rts.relayer = New(
		zerolog.Nop(),
		client.RelayerClient{},
		"",
		100,
		5,
		10,
		0,
		0,
		2,
		true,
		1*time.Second,
		1*time.Second,
		0,
		0,
		0,
		AutoRestartConfig{
			AutoRestart: false,
			Denom:       "",
			SkipError:   false,
		}, nil, []string{""},
	)
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
		types.NewDecCoinFromDec("atom", types.MustNewDecFromStr("1.13456789")),
		types.NewDecCoinFromDec("umee", types.MustNewDecFromStr("1.23456789")),
		types.NewDecCoinFromDec("juno", types.MustNewDecFromStr("1.33456789")),
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
	}

	for _, tc := range testCases {
		rts.Run(
			tc.tc, func() {
				msg, err := genRateMsgData(tc.forceRelay, tc.msgType, 0, 0, exchangeRates)
				rts.Require().NoError(err)

				var expectedMsg map[string]Msg
				err = json.Unmarshal(msg, &expectedMsg)
				rts.Require().NoError(err)

				msgKey := tc.msgType.String()
				if tc.forceRelay {
					msgKey = fmt.Sprintf("force_%s", tc.msgType.String())
				}

				rates := expectedMsg[msgKey].SymbolRates
				rts.Require().NotZero(len(rates))
				for i, rate := range rates {
					rts.Require().Equal(rate[0], exchangeRates[i].Denom)
					rts.Require().Equal(rate[1], exchangeRates[i].Amount.Mul(RateFactor).TruncateInt().String())
				}
			},
		)
	}
}

func (rts *RelayerTestSuite) Test_generateMedianRelayMsg() {
	var exchangeRates types.DecCoins
	rateMap := map[string][]interface{}{}
	for _, denom := range []string{"atom", "umee", "juno"} {
		for i := 0; i < 10; i++ {
			price := rand.Float64()
			if i%2 == 1 {
				// to have prices above 1
				price = price * 100000
			}

			priceDec := types.MustNewDecFromStr(strconv.FormatFloat(price, 'f', 9, 64))
			exchangeRates = append(exchangeRates, types.NewDecCoinFromDec(denom, priceDec))
			rateMap[denom] = append(rateMap[denom], priceDec.Mul(RateFactor).TruncateInt().String())
		}
	}

	testCases := []struct {
		tc         string
		forceRelay bool
		msgType    MsgType
	}{
		{
			tc:         "Median Relay msg",
			forceRelay: false,
			msgType:    RelayHistoricalMedian,
		},
		{
			tc:         "Median Force Relay msg",
			forceRelay: true,
			msgType:    RelayHistoricalMedian,
		},
		{
			tc:         "Deviation Relay msg",
			forceRelay: false,
			msgType:    RelayHistoricalDeviation,
		},
		{
			tc:         "Deviation Force Relay msg",
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

			key := tc.msgType.String()
			if tc.forceRelay {
				key = fmt.Sprintf("force_%s", key)
			}

			rates := expectedMsg[key].SymbolRates
			rts.Require().Len(rates, 3)

			for _, rate := range rates {
				rts.Require().Equal(rate[1], rateMap[rate[0].(string)])
			}
		})
	}
}

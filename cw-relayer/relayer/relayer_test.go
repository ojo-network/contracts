package relayer

import (
	"math/big"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"

	"github.com/ojo-network/cw-relayer/relayer/client"
	"github.com/ojo-network/cw-relayer/tools"
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
		10,
		1*time.Second,
		0,
		0,
		0,
		AutoRestartConfig{
			AutoRestart: false,
			Denom:       "",
			SkipError:   false,
		}, nil, []string{""})
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

func (rts *RelayerTestSuite) Test_generateRelayMsgs() {
	rates := types.DecCoins{
		types.NewDecCoinFromDec("atom", types.MustNewDecFromStr("1.13456789")),
		types.NewDecCoinFromDec("umee", types.MustNewDecFromStr("1.23456789")),
		types.NewDecCoinFromDec("juno", types.MustNewDecFromStr("1.33456789")),
	}

	rateMap := map[[32]byte][]*big.Int{}
	for _, rate := range rates {
		byteArray := tools.StringToByte32(rate.Denom)
		rateMap[byteArray] = append(rateMap[byteArray], decTofactorBigInt(rate.Amount))
	}

	rts.relayer.exchangeRates = rates
	rts.relayer.historicalDeviations = rates
	rts.relayer.historicalMedians = rates

	msgData := rts.relayer.genRateMsgs(0, 0)
	rts.Require().IsType(msgData, []client.PriceFeedData{})
	rts.Require().Len(msgData, rts.relayer.exchangeRates.Len())

	// since similar exchange rates are used for deviations, the value should be the same
	deviationData := rts.relayer.genDeviationsMsg(0, 0)
	rts.Require().EqualValues(msgData, deviationData)

	for i, msg := range msgData {
		rts.Require().EqualValues(msg.ResolveTime.Int64(), 0)
		rts.Require().EqualValues(msg.Id.Int64(), 0)
		rts.Require().EqualValues(msg.Value, decTofactorBigInt(rts.relayer.exchangeRates[i].Amount))
	}

	medianData := rts.relayer.genMedianMsg(0, 0)
	rts.Require().IsType(medianData, []client.PriceFeedMedianData{})
	rts.Require().Len(msgData, rts.relayer.historicalMedians.Len())

	for _, msg := range medianData {
		rts.Require().EqualValues(msg.ResolveTime.Int64(), 0)
		rts.Require().EqualValues(msg.ResolveTime.Int64(), 0)
		rts.Require().EqualValues(msg.Values, rateMap[msg.AssetName])
	}
}

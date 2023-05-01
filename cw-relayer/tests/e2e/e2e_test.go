package e2e

import (
	"time"

	"github.com/ojo-network/cw-relayer/relayer"
	"github.com/ojo-network/cw-relayer/relayer/client"
	"github.com/ojo-network/cw-relayer/tools"
)

func (s *IntegrationTestSuite) TestQueryRates() {
	var (
		rate          client.PriceFeedData
		deviationRate client.PriceFeedData
		medianRate    client.PriceFeedMedianData
		err           error
	)

	mockPrices := s.priceServer.GetMockPrices()
	s.Require().NotZero(len(mockPrices))

	// eventually the contract will have price, deviation and median data
	checkDenom := tools.StringToByte32(mockPrices[0].Denom)
	s.Require().Eventually(func() bool {
		rate, err = s.session.GetPriceData(checkDenom)
		s.Require().NoError(err)

		deviationRate, err = s.session.GetDeviationData(checkDenom)
		s.Require().NoError(err)

		medianRate, err = s.session.GetMedianData(checkDenom)
		s.Require().NoError(err)

		if rate.Id.Int64() != 0 && deviationRate.Id.Int64() != 0 && medianRate.Id.Int64() != 0 {
			return true
		}

		return false
	}, 5*time.Minute, 10*time.Second)

	// check individually for all assets
	for _, asset := range mockPrices {
		amount := relayer.DecTofactorBigInt(asset.Amount).Int64()
		checkDenom := tools.StringToByte32(asset.Denom)

		rate, err = s.session.GetPriceData(checkDenom)
		s.Require().NoError(err)
		s.Require().Equal(rate.Value.Int64(), amount)

		deviationRate, err = s.session.GetDeviationData(checkDenom)
		s.Require().NoError(err)
		s.Require().Equal(deviationRate.Value.Int64(), amount)

		medianRate, err = s.session.GetMedianData(checkDenom)
		s.Require().NoError(err)
		s.Require().Len(medianRate.Values, 1)
		s.Require().Equal(medianRate.Values[0].Int64(), amount)
	}
}

func (s *IntegrationTestSuite) TestQueryBulkRates() {
	mockPrices := s.priceServer.GetMockPrices()
	s.Require().NotZero(len(mockPrices))

	// check bulk queries
	var assetNames [][32]byte
	for _, assets := range mockPrices {
		assetNames = append(assetNames, tools.StringToByte32(assets.Denom))
	}

	var (
		rates          []client.PriceFeedData
		deviationRates []client.PriceFeedData
		medianRates    []client.PriceFeedMedianData
		err            error
	)

	s.Require().Eventually(func() bool {
		rates, err = s.session.GetPriceDataBulk(assetNames)
		s.Require().NoError(err)
		s.Require().Len(rates, len(mockPrices))

		deviationRates, err = s.session.GetDeviationDataBulk(assetNames)
		s.Require().NoError(err)
		s.Require().Len(deviationRates, len(mockPrices))

		medianRates, err = s.session.GetMedianDataBulk(assetNames)
		s.Require().NoError(err)
		s.Require().Len(medianRates, len(mockPrices))

		if rates[0].Id.Int64() != 0 && deviationRates[0].Id.Int64() != 0 && medianRates[0].Id.Int64() != 0 {
			return true
		}

		return false
	}, 5*time.Minute, 10*time.Second)

	// check individually for all assets
	for i, asset := range mockPrices {
		amount := relayer.DecTofactorBigInt(asset.Amount).Int64()

		s.Require().Equal(rates[i].Value.Int64(), amount)
		s.Require().Equal(deviationRates[i].Value.Int64(), amount)
		s.Require().Equal(medianRates[i].Values[0].Int64(), amount)
	}
}

package e2e

import (
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ojo-network/cw-relayer/relayer"
	"github.com/ojo-network/cw-relayer/relayer/client"
	"github.com/ojo-network/cw-relayer/tests/e2e/orchestrator"
	"github.com/ojo-network/cw-relayer/tools"
)

func (s *IntegrationTestSuite) TestQueryRates() {
	address := common.HexToAddress(orchestrator.ContractAddress)
	ethClient, err := ethclient.Dial(s.orchestrator.EVMRpc)
	s.Require().NoError(err)

	oracle, err := client.NewOracle(address, ethClient)
	s.Require().NoError(err)

	mockPrices := s.priceServer.GetMockPrices()
	s.Require().NotZero(len(mockPrices))

	callOpts := bind.CallOpts{
		Pending: false,
	}

	session := &client.OracleSession{
		Contract: oracle,
		CallOpts: callOpts,
	}

	// handle delay in deployment of contract
	checkDenom := tools.StringToByte32(mockPrices[0].Denom)
	_, err = oracle.GetPriceData(&callOpts, checkDenom)
	if err != nil {
		if err == bind.ErrNoCode {
			// wait till contract is deployed
			s.Require().Eventually(func() bool {
				_, err = oracle.GetPriceData(&callOpts, checkDenom)
				return err == nil
			}, 2*time.Minute, 10*time.Second)
		} else {
			s.Require().FailNow(err.Error())
		}
	}

	// eventually the contract will have price, deviation and median data
	s.Require().Eventually(func() bool {
		rate, err := oracle.GetPriceData(&callOpts, checkDenom)
		s.Require().NoError(err)

		deviationRate, err := session.GetDeviationData(checkDenom)
		s.Require().NoError(err)

		medianRate, err := session.GetMedianData(checkDenom)
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

		rate, err := oracle.GetPriceData(&callOpts, checkDenom)
		s.Require().NoError(err)
		s.Require().Equal(rate.Value.Int64(), amount)

		deviationRate, err := session.GetDeviationData(checkDenom)
		s.Require().NoError(err)
		s.Require().Equal(deviationRate.Value.Int64(), amount)

		medianRate, err := session.GetMedianData(checkDenom)
		s.Require().NoError(err)
		s.Require().Len(medianRate.Values, 1)
		s.Require().Equal(medianRate.Values[0].Int64(), amount)
	}
}

func (s *IntegrationTestSuite) TestQueryBulkRates() {
	address := common.HexToAddress(orchestrator.ContractAddress)
	ethClient, err := ethclient.Dial(s.orchestrator.EVMRpc)
	s.Require().NoError(err)

	oracle, err := client.NewOracle(address, ethClient)
	s.Require().NoError(err)

	mockPrices := s.priceServer.GetMockPrices()
	s.Require().NotZero(len(mockPrices))

	callOpts := bind.CallOpts{
		Pending: false,
	}

	session := &client.OracleSession{
		Contract: oracle,
		CallOpts: callOpts,
	}

	// handle delay in deployment of contract
	checkDenom := tools.StringToByte32(mockPrices[0].Denom)
	_, err = oracle.GetPriceData(&callOpts, checkDenom)
	if err != nil {
		if err == bind.ErrNoCode {
			// wait till contract is deployed
			s.Require().Eventually(func() bool {
				_, err = oracle.GetPriceData(&callOpts, checkDenom)
				return err == nil
			}, 2*time.Minute, 10*time.Second)
		} else {
			s.Require().FailNow(err.Error())
		}
	}

	// check bulk queries
	var assetNames [][32]byte
	for _, assets := range mockPrices {
		assetNames = append(assetNames, tools.StringToByte32(assets.Denom))
	}

	var (
		rates          []client.PriceFeedData
		deviationRates []client.PriceFeedData
		medianRates    []client.PriceFeedMedianData
	)

	s.Require().Eventually(func() bool {
		rates, err = oracle.GetPriceDataBulk(&callOpts, assetNames)
		s.Require().NoError(err)
		s.Require().Len(rates, len(mockPrices))

		deviationRates, err = session.GetDeviationDataBulk(assetNames)
		s.Require().NoError(err)
		s.Require().Len(deviationRates, len(mockPrices))

		medianRates, err = session.GetMedianDataBulk(assetNames)
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

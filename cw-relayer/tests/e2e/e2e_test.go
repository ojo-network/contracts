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

func (s *IntegrationTestSuite) TestQueryRateAndReferenceData() {
	//time.Sleep(5 * time.Minute)
	address := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")
	ethClient, err := ethclient.Dial(orchestrator.EVMRpc)
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

	// eventually the contract will have price, deviation and median data
	s.Require().Eventually(func() bool {
		rate, err := oracle.GetPriceData(&callOpts, tools.StringToByte32(mockPrices[0].Denom))
		s.Require().NoError(err)

		deviationRate, err := session.GetDeviationData(tools.StringToByte32(mockPrices[0].Denom))
		s.Require().NoError(err)

		medianRate, err := session.GetMedianData(tools.StringToByte32(mockPrices[0].Denom))
		s.Require().NoError(err)

		if rate.Id.Int64() != 0 && deviationRate.Id.Int64() != 0 && medianRate.Id.Int64() != 0 {
			return true
		}

		return false
	}, 4*time.Minute, 10*time.Second)

	// check individually for all assets
	for _, asset := range mockPrices {
		rate, err := oracle.GetPriceData(&callOpts, tools.StringToByte32(asset.Denom))
		s.Require().NoError(err)
		s.Require().Equal(rate.Value.Int64(), asset.Amount.Mul(relayer.RateFactor).TruncateInt().Int64())

		deviationRate, err := session.GetDeviationData(tools.StringToByte32(asset.Denom))
		s.Require().NoError(err)
		s.Require().Equal(deviationRate.Value.Int64(), asset.Amount.Mul(relayer.RateFactor).TruncateInt().Int64())

		medianRate, err := session.GetMedianData(tools.StringToByte32(asset.Denom))
		s.Require().NoError(err)
		s.Require().Len(medianRate.Values, 1)
		s.Require().EqualValues(medianRate.Values[0].Int64(), asset.Amount.Mul(relayer.RateFactor).TruncateInt().Int64())
	}
}

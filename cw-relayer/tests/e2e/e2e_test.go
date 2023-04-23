package e2e

import (
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ojo-network/cw-relayer/relayer/client"
	"github.com/ojo-network/cw-relayer/tools"
)

type ()

const testConfigTimeout = 2 * time.Minute

var (

	// used to convert rate from reference data queries to USD
	refDataFactor = types.NewDec(10).Power(18)
)

func (s *IntegrationTestSuite) TestQueryRateAndReferenceData() {
	address := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")
	ethClient, err := ethclient.Dial(s.orchestrator.WasmRPC)
	s.Require().NoError(err)

	oracle, err := client.NewOracle(address, ethClient)
	s.Require().NoError(err)
	callOpts := bind.CallOpts{
		Pending: false,
	}
	mockPrices := s.priceServer.GetMockPrices()
	s.T().Log(mockPrices)

	session := &client.OracleSession{
		Contract: oracle,
		CallOpts: bind.CallOpts{
			Pending: false,
		},
	}

	_, err = client.NewOracleCaller(common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3"), ethClient)
	s.Require().NoError(err)

	s.Require().Eventually(func() bool {

		getprices, err := oracle.GetPriceData(&callOpts, tools.StringToByte32(mockPrices[0].Denom))
		s.Require().NoError(err)

		s.T().Log(string(getprices.AssetName[:]))

		pricedata, err := session.GetDeviationData(tools.StringToByte32(mockPrices[2].Denom))
		s.Require().NoError(err)

		medianData, err := session.GetMedianData(tools.StringToByte32(mockPrices[2].Denom))
		s.Require().NoError(err)
		s.T().Log(string(medianData.AssetName[:]))

		priceData, err := session.GetPriceData(tools.StringToByte32(mockPrices[1].Denom))
		s.Require().NoError(err)
		s.T().Log(string(priceData.AssetName[:]))

		if pricedata.Value.Int64() != 0 || priceData.Value.Int64() != 0 {
			return false
		}

		return false
	}, 4*time.Minute, 10*time.Second)
}

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

	//
	//time.Sleep(1 * time.Minute)

	s.Require().Eventually(func() bool {

		getprices, err := oracle.GetPriceData(&callOpts, tools.StringToByte32(mockPrices[0].Denom))
		s.Require().NoError(err)

		s.T().Log(getprices)

		pricedata, err := session.GetDeviationData(tools.StringToByte32(mockPrices[2].Denom))
		s.Require().NoError(err)
		s.T().Log(pricedata)

		medianData, err := session.GetMedianData(tools.StringToByte32(mockPrices[2].Denom))
		s.Require().NoError(err)
		s.T().Log(medianData)

		priceData, err := session.GetPriceData(tools.StringToByte32(mockPrices[1].Denom))
		s.Require().NoError(err)
		s.T().Log(priceData)

		if pricedata.Value.Int64() != 0 || priceData.Value.Int64() != 0 {
			return true
		}

		return false
	}, 3*time.Minute, 10*time.Second)
}

//}

//
//func (s *IntegrationTestSuite) TestQueryReferenceDataBulk() {
//	grpcConn, err := grpc.Dial(
//		s.orchestrator.QueryRpc,
//		grpc.WithTransportCredentials(insecure.NewCredentials()),
//	)
//
//	s.Require().NoError(err)
//	defer grpcConn.Close()
//
//	mockPrices := s.priceServer.GetMockPrices()
//
//	testCases := []struct {
//		tc      string
//		prepare func() ([]byte, error)
//		factor  types.Dec
//	}{
//		{
//			tc: "query reference data in bulk",
//			prepare: func() ([]byte, error) {
//				var symbolData [][2]string
//				for _, mockPrice := range mockPrices {
//					symbolData = append(symbolData, [2]string{mockPrice.Denom, "USD"})
//				}
//
//				msg := refMsgBulk{symbolPairs{SymbolPairs: symbolData}}
//				data, err := json.Marshal(msg)
//				if err != nil {
//					return nil, err
//				}
//
//				return data, nil
//			},
//			factor: refDataFactor,
//		},
//		{
//			tc: "query deviation data in bulk",
//			prepare: func() ([]byte, error) {
//				var denoms []string
//				for _, mockPrice := range mockPrices {
//					denoms = append(denoms, mockPrice.Denom)
//				}
//
//				msg := deviationRateMsgBulk{symbols{Symbols: denoms}}
//				data, err := json.Marshal(msg)
//				if err != nil {
//					return nil, err
//				}
//
//				return data, nil
//			},
//			factor: relayer.RateFactor,
//		},
//	}
//
//	for _, tc := range testCases {
//		s.Run(tc.tc, func() {
//			queryClient := wasmtypes.NewQueryClient(grpcConn)
//			data, err := tc.prepare()
//			s.Require().NoError(err)
//
//			query := wasmtypes.QuerySmartContractStateRequest{
//				Address:   s.orchestrator.ContractAddress,
//				QueryData: data,
//			}
//
//			ctx, cancel := context.WithTimeout(context.Background(), testConfigTimeout)
//			defer cancel()
//
//			s.Require().Eventually(func() bool {
//				queryResponse, err := queryClient.SmartContractState(ctx, &query)
//				if err != nil {
//					return false
//				}
//
//				if queryResponse != nil {
//					var resp []map[string]string
//					err = json.Unmarshal(queryResponse.Data, &resp)
//					if err != nil {
//						return false
//					}
//
//					for i, respData := range resp {
//						s.Require().Equal(respData["rate"], mockPrices[i].Amount.Mul(tc.factor).TruncateInt().String())
//					}
//
//					return true
//				}
//
//				return false
//			},
//				1*time.Minute,
//				time.Second*4,
//				"failed to query prices from contract",
//			)
//		})
//	}
//}
//
//func (s *IntegrationTestSuite) TestQueryMedianRates() {
//	grpcConn, err := grpc.Dial(
//		s.orchestrator.QueryRpc,
//		grpc.WithTransportCredentials(insecure.NewCredentials()),
//	)
//
//	s.Require().NoError(err)
//	defer grpcConn.Close()
//
//	mockPrices := s.priceServer.GetMockPrices()
//
//	testCases := []struct {
//		tc      string
//		prepare func() ([]byte, error)
//		factor  types.Dec
//		bulk    bool
//	}{
//		{
//			tc: "query median rate data from contract",
//			prepare: func() ([]byte, error) {
//
//				msg := medianRateMsg{Ref: symbol{mockPrices[0].Denom}}
//				data, err := json.Marshal(msg)
//				if err != nil {
//					return nil, err
//				}
//
//				return data, nil
//			},
//			factor: relayer.RateFactor,
//			bulk:   false,
//		},
//		{
//			tc: "query median ref data bulk",
//			prepare: func() ([]byte, error) {
//				var denoms []string
//				for _, mockPrice := range mockPrices {
//					denoms = append(denoms, mockPrice.Denom)
//				}
//
//				msg := medianRefMsgBulk{symbols{denoms}}
//				data, err := json.Marshal(msg)
//				if err != nil {
//					return nil, err
//				}
//
//				return data, nil
//			},
//			factor: relayer.RateFactor,
//			bulk:   true,
//		},
//	}
//
//	for _, tc := range testCases {
//		s.Run(tc.tc, func() {
//			queryClient := wasmtypes.NewQueryClient(grpcConn)
//			data, err := tc.prepare()
//			s.Require().NoError(err)
//
//			query := wasmtypes.QuerySmartContractStateRequest{
//				Address:   s.orchestrator.ContractAddress,
//				QueryData: data,
//			}
//
//			ctx, cancel := context.WithTimeout(context.Background(), testConfigTimeout)
//			defer cancel()
//
//			s.Require().Eventually(func() bool {
//				queryResponse, err := queryClient.SmartContractState(ctx, &query)
//				if err != nil {
//					return false
//				}
//
//				if queryResponse != nil {
//					if tc.bulk {
//						var resp []map[string]interface{}
//						err = json.Unmarshal(queryResponse.Data, &resp)
//						if err != nil {
//							return false
//						}
//
//						for i, respData := range resp {
//							s.Require().Equal(respData["rates"].([]interface{})[0], mockPrices[i].Amount.Mul(tc.factor).TruncateInt().String())
//						}
//					} else {
//						var resp map[string]interface{}
//						err = json.Unmarshal(queryResponse.Data, &resp)
//						if err != nil {
//							return false
//						}
//
//						s.Require().Equal(resp["rates"].([]interface{})[0], mockPrices[0].Amount.Mul(tc.factor).TruncateInt().String())
//					}
//
//					return true
//				}
//
//				return false
//			},
//				2*time.Minute,
//				time.Second*4,
//				"failed to query prices from contract",
//			)
//		})
//	}
//}

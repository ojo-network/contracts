package e2e

import (
	"context"
	"encoding/json"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ojo-network/cw-relayer/relayer"
)

type (
	rateMsg struct {
		Ref symbol `json:"get_ref"`
	}

	symbol struct {
		Symbol string `json:"symbol"`
	}

	refMsg struct {
		RefData symbolPair `json:"get_reference_data"`
	}

	symbolPair struct {
		SymbolPair [2]string `json:"symbol_pair"`
	}

	refMsgBulk struct {
		RefData symbolPairs `json:"get_reference_data_bulk"`
	}

	symbolPairs struct {
		SymbolPairs [][2]string `json:"symbol_pairs"`
	}
)

var (
	// only used when querying prices with reference to other tokens
	contractFactor = relayer.Factor.Power(2)
)

func (s *IntegrationTestSuite) TestQueryRateAndReferenceData() {
	grpcConn, err := grpc.Dial(
		s.orchestrator.QueryRpc,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	s.Require().NoError(err)
	defer grpcConn.Close()

	mockPrices := s.priceServer.GetMockPrices()

	testCases := []struct {
		tc      string
		prepare func() ([]byte, error)
		rate    string
	}{
		{
			tc: "query rate from contract",
			prepare: func() ([]byte, error) {
				msg := rateMsg{Ref: symbol{Symbol: mockPrices[0].Denom}}
				data, err := json.Marshal(msg)
				if err != nil {
					return nil, err
				}

				return data, err
			},
			rate: mockPrices[0].Amount.Mul(relayer.Factor).TruncateInt().String(),
		},
		{
			tc: "query reference data in USD from contract",
			prepare: func() ([]byte, error) {
				msg := refMsg{symbolPair{SymbolPair: [2]string{mockPrices[0].Denom, "USD"}}}
				data, err := json.Marshal(msg)
				if err != nil {
					return nil, err
				}

				return data, nil
			},
			rate: mockPrices[0].Amount.Mul(contractFactor).TruncateInt().String(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.tc, func() {
			queryClient := wasmtypes.NewQueryClient(grpcConn)
			data, err := tc.prepare()
			s.Require().NoError(err)

			query := wasmtypes.QuerySmartContractStateRequest{
				Address:   s.orchestrator.ContractAddress,
				QueryData: data,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			s.Require().Eventually(func() bool {
				queryResponse, err := queryClient.SmartContractState(ctx, &query)
				if err != nil {
					return false
				}

				if queryResponse != nil {
					resp := map[string]string{}
					err = json.Unmarshal(queryResponse.Data, &resp)
					if err != nil {
						return false
					}

					s.Require().Equal(resp["rate"], tc.rate)

					return true
				}

				return false
			},
				1*time.Minute,
				time.Second*4,
				"failed to query prices from contract",
			)
		})
	}
}

func (s *IntegrationTestSuite) TestQueryReferenceDataBulk() {
	grpcConn, err := grpc.Dial(
		s.orchestrator.QueryRpc,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	s.Require().NoError(err)
	defer grpcConn.Close()

	mockPrices := s.priceServer.GetMockPrices()

	testCases := []struct {
		tc      string
		prepare func() ([]byte, error)
	}{
		{
			tc: "query reference data in bulk",
			prepare: func() ([]byte, error) {
				var symbolData [][2]string
				for _, mockPrice := range mockPrices {
					symbolData = append(symbolData, [2]string{mockPrice.Denom, "USD"})
				}

				msg := refMsgBulk{symbolPairs{SymbolPairs: symbolData}}
				data, err := json.Marshal(msg)
				if err != nil {
					return nil, err
				}

				return data, nil
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.tc, func() {
			queryClient := wasmtypes.NewQueryClient(grpcConn)
			data, err := tc.prepare()
			s.Require().NoError(err)

			query := wasmtypes.QuerySmartContractStateRequest{
				Address:   s.orchestrator.ContractAddress,
				QueryData: data,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			s.Require().Eventually(func() bool {
				queryResponse, err := queryClient.SmartContractState(ctx, &query)
				if err != nil {
					return false
				}

				if queryResponse != nil {
					var resp []map[string]string
					err = json.Unmarshal(queryResponse.Data, &resp)
					if err != nil {
						return false
					}

					for i, respData := range resp {
						s.Require().Equal(respData["rate"], mockPrices[i].Amount.Mul(contractFactor).TruncateInt().String())
					}

					return true
				}

				return false
			},
				1*time.Minute,
				time.Second*4,
				"failed to query prices from contract",
			)
		})
	}
}

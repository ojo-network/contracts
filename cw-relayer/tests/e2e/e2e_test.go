package e2e

import (
	"context"
	"encoding/json"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type (
	RateMsg struct {
		Ref Rate `json:"get_ref"`
	}

	Rate struct {
		Symbol string `json:"symbol"`
	}

	RateData struct {
		Rate        string `json:"rate"`
		ResolveTime string `json:"resolve_time"`
		RequestId   string `json:"request_id"`
	}

	RefMsg struct {
		RefData Symbol `json:"get_reference_data"`
	}

	Symbol struct {
		SymbolPair [2]string `json:"symbol_pair"`
	}

	RefData struct {
		Rate             string `json:"rate"`
		LastUpdatedBase  string `json:"last_updated_base"`
		LastUpdatedQuote string `json:"last_updated_quote"`
	}

	RefMsgBulk struct {
		RefData Symbols `json:"get_reference_data_bulk"`
	}

	Symbols struct {
		SymbolPairs [][2]string `json:"symbol_pairs"`
	}
)

func (s *IntegrationTestSuite) TestQueryFromContract() {
	grpcConn, err := grpc.Dial(
		s.orchestrator.QueryRpc,
		// the Cosmos SDK doesn't support any transport security mechanism
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	s.Require().NoError(err)
	defer grpcConn.Close()

	testCases := []struct {
		tc      string
		prepare func() ([]byte, error)
		assert  func(queryData []byte) bool
	}{
		{
			tc: "query rate from contract",
			prepare: func() ([]byte, error) {
				msg := RateMsg{Ref: Rate{Symbol: "TEST"}}
				data, err := json.Marshal(msg)
				if err != nil {
					return nil, err
				}

				return data, err
			},
			assert: func(queryData []byte) bool {
				var resp RateData
				err = json.Unmarshal(queryData, &resp)
				if err != nil {
					return false
				}

				s.Require().Equal(resp.Rate, "1234567890")

				return true
			},
		},
		{
			tc: "query reference data in USD from contract",
			prepare: func() ([]byte, error) {
				msg := RefMsg{Symbol{[2]string{"TEST", "USD"}}}
				data, err := json.Marshal(msg)
				if err != nil {
					return nil, err
				}

				return data, nil
			},
			assert: func(queryData []byte) bool {
				var resp RefData
				err = json.Unmarshal(queryData, &resp)
				if err != nil {
					return false
				}

				s.Require().Equal(resp.Rate, "1234567890000000000")

				return true

			},
		},
		{
			tc: "query reference data in bulk",
			prepare: func() ([]byte, error) {
				msg := RefMsgBulk{Symbols{[][2]string{{"TEST", "USD"}, {"STAKE", "USD"}}}}
				data, err := json.Marshal(msg)
				if err != nil {
					return nil, err
				}

				return data, nil
			},
			assert: func(queryData []byte) bool {
				var resp []RefData
				err = json.Unmarshal(queryData, &resp)
				if err != nil {
					return false
				}

				s.Require().Equal(resp[0].Rate, "1234567890000000000")
				s.Require().Equal(resp[0].Rate, "1234567890000000000")

				s.Require().Equal(resp[1].Rate, "1234560000000000")
				s.Require().Equal(resp[1].Rate, "1234560000000000")

				return true

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
					return tc.assert(queryResponse.Data)
				}

				return false
			},
				2*time.Minute,
				time.Second*4,
				"failed to query prices from contract",
			)
		})
	}
}

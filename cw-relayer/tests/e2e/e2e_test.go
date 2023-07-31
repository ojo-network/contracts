package e2e

import (
	"context"
	"encoding/json"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ojo-network/cw-relayer/tests/e2e/orchestrator"
)

type (
	GetPrice struct {
		SymbolRequested Symbol `json:"get_price"`
	}

	Symbol struct {
		Symbol string `json:"symbol"`
	}

	LastPing struct {
		Relayer Relayer `json:"last_ping"`
	}

	Relayer struct {
		Relayer string `json:"relayer"`
	}
)

func (s *IntegrationTestSuite) generateQuery(symbol string) wasmtypes.QuerySmartContractStateRequest {
	data, _ := json.Marshal(GetPrice{SymbolRequested: Symbol{Symbol: symbol}})
	return wasmtypes.QuerySmartContractStateRequest{
		Address:   s.orchestrator.QueryContractAddress,
		QueryData: data,
	}
}

func (s *IntegrationTestSuite) TestCallback() {
	grpcConn, err := grpc.Dial(
		s.orchestrator.QueryRpc,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	s.Require().NoError(err)
	defer grpcConn.Close()

	queryClient := wasmtypes.NewQueryClient(grpcConn)

	// fetch last ping for the relayer
	lastPingData, err := json.Marshal(LastPing{Relayer: Relayer{Relayer: s.orchestrator.WasmChain.Address}})
	s.Require().NoError(err)

	s.Require().Eventually(
		func() bool {
			pingQuery := wasmtypes.QuerySmartContractStateRequest{Address: s.orchestrator.ContractAddress, QueryData: lastPingData}
			resp, err := queryClient.SmartContractState(context.Background(), &pingQuery)
			if err != nil {
				return false
			}

			if len(resp.String()) > 0 {
				return true
			}

			return false
		}, 2*time.Minute, 20*time.Second)

	s.Require().Eventually(
		func() bool {
			err = s.orchestrator.RequestMsg(orchestrator.Price, s.orchestrator.QueryContractAddress, "TEST-0")
			if err != nil {
				return false
			}

			query := s.generateQuery("TEST-0")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			resp, err := queryClient.SmartContractState(ctx, &query)
			if err != nil || len(resp.String()) == 0 {
				return false
			}

			var rate string
			err = json.Unmarshal(resp.Data, &rate)
			if err != nil {
				return false
			}

			if rate != "0" {
				return true
			}

			return false

		},
		7*time.Minute, 20*time.Second,
		"relayer request and callback failed")
}

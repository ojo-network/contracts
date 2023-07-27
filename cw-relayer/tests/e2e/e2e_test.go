package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type (
	GetPrice struct {
		Denom Symbol `json:"get_price"`
	}

	Symbol struct {
		Symbol string `json:"symbol"`
	}
)

var (
// used to convert rate from reference data queries to USD
// refDataFactor = types.NewDec(10).Power(18)
)

func (s *IntegrationTestSuite) TestCallback() {
	grpcConn, err := grpc.Dial(
		s.orchestrator.QueryRpc,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	s.Require().NoError(err)
	defer grpcConn.Close()

	queryClient := wasmtypes.NewQueryClient(grpcConn)
	msg := GetPrice{Denom: Symbol{Symbol: "TEST-0"}}

	data, err := json.Marshal(msg)
	s.Require().NoError(err)

	query := wasmtypes.QuerySmartContractStateRequest{
		Address:   s.orchestrator.QueryContractAddress,
		QueryData: data,
	}

	err = s.orchestrator.RequestPrices(s.orchestrator.QueryContractAddress, "TEST-0")
	s.T().Log(err)

	for i := 0; i < 100; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := queryClient.SmartContractState(ctx, &query)
		fmt.Println(err)

		fmt.Println(resp.String())

		if len(resp.String()) == 0 {

			fmt.Println(resp.String())
			fmt.Println(err)
		}
	}
}

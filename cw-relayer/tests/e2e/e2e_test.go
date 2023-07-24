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
		Request struct{} `json:"get_price"`
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
	//msg := GetPrice{Request: struct{}{}}

	data, err := json.Marshal("get_price")
	s.Require().NoError(err)

	query := wasmtypes.QuerySmartContractStateRequest{
		Address:   s.orchestrator.QueryContractAddress,
		QueryData: data,
	}

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
	s.Require().Eventually(func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := queryClient.SmartContractState(ctx, &query)
		if err != nil {
			return false
		}

		fmt.Println(resp.String())

		if len(resp.String()) == 0 {

			fmt.Println(resp.String())
			fmt.Println(err)
		}

		return false
	}, 1*time.Minute, 10*time.Second)
}

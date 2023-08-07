package e2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ojo-network/cw-relayer/tests/e2e/orchestrator"
	"github.com/ojo-network/cw-relayer/tests/e2e/server"
)

const QUERY_PORT = "9090"

type IntegrationTestSuite struct {
	suite.Suite

	orchestrator orchestrator.Orchestrator
	priceServer  server.Server
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.orchestrator = orchestrator.Orchestrator{}
	s.priceServer = server.Server{}

	s.T().Log("---> initializing docker resources")
	err := s.orchestrator.InitDockerResources(s.T())
	s.Require().NoError(err)

	s.T().Log("---> initializing mock price server")
	err = s.priceServer.InitMockPriceServer(QUERY_PORT)
	s.Require().NoError(err)

	s.T().Log("---> relayer ping check")
	grpcConn, err := grpc.Dial(
		s.orchestrator.QueryRpc,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	s.Require().NoError(err)
	defer grpcConn.Close()
	queryClient := wasmtypes.NewQueryClient(grpcConn)

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
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("---> tearing down")
	err := s.orchestrator.TearDownDockerResources()
	s.Require().NoError(err)
}

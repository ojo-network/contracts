package e2e

import (
	"testing"

	"github.com/stretchr/testify/suite"

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

	//time.Sleep(2 * time.Minute)
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
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("---> tearing down")
	//err := s.orchestrator.TearDownDockerResources()
	//s.Require().NoError(err)
}

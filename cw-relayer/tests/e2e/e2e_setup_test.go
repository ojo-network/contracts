package e2e

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ojo-network/cw-relayer/tests/e2e/orchestrator"
)

type IntegrationTestSuite struct {
	suite.Suite

	orchestrator orchestrator.Orchestrator
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.orchestrator = orchestrator.Orchestrator{}
	s.T().Log("---> initializing docker resources")
	s.Require().NoError(s.orchestrator.InitDockerResources(s.T()))
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("---> tearing down")
	s.Require().NoError(s.orchestrator.TearDownDockerResources())
}

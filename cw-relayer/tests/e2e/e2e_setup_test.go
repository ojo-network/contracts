package e2e

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/suite"

	"github.com/ojo-network/cw-relayer/relayer/client"
	"github.com/ojo-network/cw-relayer/tests/e2e/orchestrator"
	"github.com/ojo-network/cw-relayer/tests/e2e/server"
	"github.com/ojo-network/cw-relayer/tools"
)

type IntegrationTestSuite struct {
	suite.Suite

	orchestrator orchestrator.Orchestrator
	priceServer  server.Server
	session      *client.OracleSession
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
	err = s.priceServer.InitMockPriceServer()
	s.Require().NoError(err)

	ethClient, err := ethclient.Dial(s.orchestrator.EVMRpc)
	s.Require().NoError(err)

	oracle, err := client.NewOracle(orchestrator.ContractAddress, ethClient)
	s.Require().NoError(err)

	s.session = &client.OracleSession{
		Contract: oracle,
		CallOpts: bind.CallOpts{
			Pending: false,
		},
	}

	mockPrices := s.priceServer.GetMockPrices()
	s.Require().NotZero(len(mockPrices))

	s.T().Log("---> waiting for contract deployment")
	checkDenom := tools.StringToByte32(mockPrices[0].Denom)
	_, err = s.session.GetPriceData(checkDenom)
	if err != nil {
		if err == bind.ErrNoCode {
			// wait till contract is deployed
			s.Require().Eventually(func() bool {
				_, err = s.session.GetPriceData(checkDenom)
				return err == nil
			}, 2*time.Minute, 10*time.Second)
		} else {
			s.Require().FailNow(err.Error())
		}
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("---> tearing down")
	err := s.orchestrator.TearDownDockerResources()
	s.Require().NoError(err)
}

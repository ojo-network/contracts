package orchestrator

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
)

const (
	IMAGE_REPO     = "cw-relayer-evm"
	CONTAINER_NAME = "cw-relayer-evm"
)

var (
	relayerAddress  = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	privKey         = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	ContractAddress = common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")
)

// Orchestrator is responsible for managing docker resources,
// their configuration files, and environment variables.
type Orchestrator struct {
	dockerPool    *dockertest.Pool
	dockerNetwork *dockertest.Network

	evmResource *dockertest.Resource
	EVMRpc      string
	QueryRpc    string
}

func (o *Orchestrator) InitDockerResources(t *testing.T) error {
	var err error
	t.Log("-> initializing docker network")
	err = o.initNetwork()
	if err != nil {
		return err
	}

	t.Log("-> initializing relayer")
	err = o.initRelayer()
	if err != nil {
		return err
	}

	t.Log("-> verifying evm node is creating blocks")
	require.Eventually(
		t,
		func() bool {
			blockHeight, err := o.evmBlockHeight()
			if err != nil {
				return false
			}
			return blockHeight >= 3
		},
		time.Minute,
		time.Second*2,
		"hardhat node failed to produce blocks",
	)

	t.Log("--> deploy contract")
	err = o.deployContract()
	if err != nil {
		return err
	}

	t.Log("--> start relayer")
	err = o.startRelayer()
	if err != nil {
		return err
	}

	return nil
}

func (o *Orchestrator) TearDownDockerResources() error {
	err := o.dockerPool.Purge(o.evmResource)
	if err != nil {
		return err
	}

	return o.dockerPool.Client.RemoveNetwork(o.dockerNetwork.Network.ID)
}

func (o *Orchestrator) initNetwork() error {
	var err error
	o.dockerPool, err = dockertest.NewPool("")
	if err != nil {
		return err
	}

	o.dockerNetwork, err = o.dockerPool.CreateNetwork("e2e_test_network")
	if err != nil {
		return err
	}
	return nil
}

func noRestart(config *docker.HostConfig) {
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}

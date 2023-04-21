package orchestrator

import (
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
)

const (
	IMAGE_REPO     = "cw-relayer-evm"
	CONTAINER_NAME = "cw-relayer-evm"
)

// Orchestrator is responsible for managing docker resources,
// their configuration files, and environment variables.
type Orchestrator struct {
	dockerPool    *dockertest.Pool
	dockerNetwork *dockertest.Network

	wasmdResource *dockertest.Resource
	WasmRPC       string

	QueryRpc        string
	ContractAddress string
}

func (o *Orchestrator) InitDockerResources(t *testing.T) error {
	var err error

	o.WasmRPC = "http://localhost:8545"

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

	t.Log("-> verifying wasm node is creating blocks")
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

	return nil
}

func (o *Orchestrator) TearDownDockerResources() error {
	err := o.dockerPool.Purge(o.wasmdResource)
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

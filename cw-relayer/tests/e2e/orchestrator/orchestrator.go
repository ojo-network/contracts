package orchestrator

import (
	"container/list"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
)

const (
	WASMD_IMAGE_REPO     = "cw-relayer"
	WASMD_CONTAINER_NAME = "cw-relayer"
	WASMD_TRPC_PORT      = "26657"
	WASM_GRPC_PORT       = "8080"
	QUERY_GRPC_PORT      = "9090"
)

// Orchestrator is responsible for managing docker resources,
// their configuration files, and environment variables.
type Orchestrator struct {
	dockerPool    *dockertest.Pool
	dockerNetwork *dockertest.Network

	wasmdResource *dockertest.Resource
	wasmRPC       *rpchttp.HTTP
	wasmChain     *Chain

	wsConn       *websocket.Conn
	wsCancelFunc context.CancelFunc
	wsMessages   *list.List
	wsMutex      sync.RWMutex

	QueryRpc        string
	ContractAddress string
}

func (o *Orchestrator) InitDockerResources(t *testing.T) error {
	var err error

	t.Log("-> initializing docker network")
	err = o.initNetwork()
	if err != nil {
		return err
	}

	t.Log("-> initializing wasm node")
	err = o.initWasmd()
	if err != nil {
		return err
	}

	t.Log("-> verifying wasm node is creating blocks")
	require.Eventually(
		t,
		func() bool {
			blockHeight, err := o.wasmdBlockHeight()
			if err != nil {
				return false
			}
			return blockHeight >= 3
		},
		time.Minute,
		time.Second*2,
		"wasmd node failed to produce blocks",
	)

	t.Log("-> initializing mock price server")
	err = o.InitMockPriceServer(QUERY_GRPC_PORT)
	if err != nil {
		return err
	}

	t.Log("-> initializing wasm contract")
	err = o.deployAndInitContract()
	if err != nil {
		return err
	}

	t.Log("-> fetching wasm contract address")
	err = o.setContractAddress()
	if err != nil {
		return err
	}

	t.Log("-> adding Relayer to contract")
	err = o.addRelayerToContract(o.ContractAddress, o.wasmChain.address)
	if err != nil {
		return err
	}

	t.Log("-> starting Relayer")
	err = o.startRelayer(o.ContractAddress)
	if err != nil {
		return err
	}

	return nil
}

func (o *Orchestrator) TearDownDockerResources() error {
	var err error

	err = o.dockerPool.Purge(o.wasmdResource)
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

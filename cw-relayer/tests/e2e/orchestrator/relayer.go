package orchestrator

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func (o *Orchestrator) initRelayer() error {
	var err error

	o.wasmdResource, err = o.dockerPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:         CONTAINER_NAME,
			Repository:   IMAGE_REPO,
			NetworkID:    o.dockerNetwork.Network.ID,
			ExtraHosts:   []string{"host.docker.internal:host-gateway"},
			PortBindings: map[docker.Port][]docker.PortBinding{},
			Env: []string{
				fmt.Sprintf("EVM_ADDRESS=%s", RelayerAddress),
				fmt.Sprintf("EVM_PRIV_KEY=%s", priv_key),
			},
			//Mounts: []string{fmt.Sprintf("%s:/config", configDir)},
			Entrypoint: []string{
				"cw-relayer",
				"/usr/local/config.toml",
			},
		},
		noRestart,
	)

	return err
}

func (o *Orchestrator) evmBlockHeight() (uint64, error) {
	conn, err := ethclient.Dial(EVMRpc)
	if err != nil {
		return 0, err
	}

	height, err := conn.BlockNumber(context.Background())
	if err != nil {
		return 0, err

	}
	return height, nil
}

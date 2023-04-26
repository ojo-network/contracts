package orchestrator

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/ojo-network/cw-relayer/tests/e2e/server"
)

func (o *Orchestrator) initRelayer() error {
	var err error

	o.evmResource, err = o.dockerPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:         CONTAINER_NAME,
			Repository:   IMAGE_REPO,
			NetworkID:    o.dockerNetwork.Network.ID,
			ExtraHosts:   []string{"host.docker.internal:host-gateway"},
			ExposedPorts: []string{"8545"},
			PortBindings: map[docker.Port][]docker.PortBinding{},
			Env: []string{
				fmt.Sprintf("EVM_RELAYER_ADDRESS=%s", RelayerAddress),
				fmt.Sprintf("EVM_PRIV_KEY=%s", priv_key),
				fmt.Sprintf("EVM_CONTRACT_ADDRESS=%s", ContractAddress),
				fmt.Sprintf("EVM_QUERY_RPC=%s", fmt.Sprintf("host.docker.internal:%v", server.QUERY_PORT)),
			},
		},
		noRestart,
	)

	if err != nil {
		return err
	}

	return o.setEndpoint()
}

func (o *Orchestrator) setEndpoint() error {
	path := o.evmResource.GetHostPort(fmt.Sprintf("%s/tcp", "8545"))
	if path == "" {
		return fmt.Errorf("resource path empty")
	}
	o.EVMRpc = fmt.Sprintf("http://%s", path)

	return nil
}

func (o *Orchestrator) evmBlockHeight() (uint64, error) {
	conn, err := ethclient.Dial(o.EVMRpc)
	if err != nil {
		return 0, err
	}

	height, err := conn.BlockNumber(context.Background())
	if err != nil {
		return 0, err

	}
	return height, nil
}

func (o *Orchestrator) execEvm(command []string) (err error) {
	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	exec, err := o.dockerPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      context.Background(),
		AttachStdout: true,
		AttachStderr: true,
		Container:    o.evmResource.Container.ID,
		User:         "root",
		Cmd:          command,
	})

	if err != nil {
		return
	}

	err = o.dockerPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      context.Background(),
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})

	if err != nil {
		return
	}

	errOutput := errBuf.String()
	if len(errOutput) > 0 {
		if strings.Contains(errOutput, "gas estimate") {
			return nil
		}
		err = fmt.Errorf("error executing command %s, error %s", strings.Join(command, " "), errOutput)
	}

	return err
}

func (o *Orchestrator) deployContract() error {
	return o.execEvm(
		[]string{
			"sh", "-c", "yarn hardhat run ./scripts/deploy.ts --network localhost > ./deploy.log 2>&1 &",
		},
	)
}

func (o *Orchestrator) startRelayer() error {
	return o.execEvm(
		[]string{
			"sh", "-c", "chmod +x ./relayer_bootstrap.sh && ./relayer_bootstrap.sh",
		})
}

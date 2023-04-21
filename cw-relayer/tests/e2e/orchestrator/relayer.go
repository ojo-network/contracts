package orchestrator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func (o *Orchestrator) initRelayer() error {
	var err error

	configDir, err := o.initConfigs()
	if err != nil {
		return err
	}

	o.wasmdResource, err = o.dockerPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:         CONTAINER_NAME,
			Repository:   IMAGE_REPO,
			NetworkID:    o.dockerNetwork.Network.ID,
			ExtraHosts:   []string{"host.docker.internal:host-gateway"},
			PortBindings: map[docker.Port][]docker.PortBinding{},
			Mounts:       []string{fmt.Sprintf("%s:/config", configDir)},
			Entrypoint: []string{
				"cw-relayer",
				"config/relayer-config.toml",
			},
		},
		noRestart,
	)

	return err
}

func (o *Orchestrator) initConfigs() (dir string, err error) {
	dir, err = os.MkdirTemp("", "e2e-configs")
	if err != nil {
		return
	}

	err = copyFiles("./config/", dir)
	if err != nil {
		return
	}

	return
}

func (o *Orchestrator) evmBlockHeight() (uint64, error) {
	conn, err := ethclient.Dial(o.WasmRPC)
	if err != nil {
		return 0, err
	}

	height, err := conn.BlockNumber(context.Background())
	if err != nil {
		return 0, err

	}
	return height, nil
}

func (o *Orchestrator) execWasmCmd(command []string) (err error) {
	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	exec, err := o.dockerPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      context.Background(),
		AttachStdout: true,
		AttachStderr: true,
		Container:    o.wasmdResource.Container.ID,
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
		err = fmt.Errorf("error executing command %s", strings.Join(command, " "))
	}

	return err
}

package orchestrator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	tmconfig "github.com/tendermint/tendermint/config"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
)

func (o *Orchestrator) initWasmd() error {
	var err error

	configDir, err := o.initWasmConfigs()
	if err != nil {
		return err
	}

	o.wasmChain = NewChain("test-wasm")

	o.wasmdResource, err = o.dockerPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       WASMD_CONTAINER_NAME,
			Repository: WASMD_IMAGE_REPO,
			NetworkID:  o.dockerNetwork.Network.ID,
			Env: []string{
				fmt.Sprintf("E2E_WASMD_CHAIN_ID=%s", o.wasmChain.chainId),
				fmt.Sprintf("E2E_WASMD_VAL_MNEMONIC=%s", o.wasmChain.val_mnemonic),
				fmt.Sprintf("E2E_WASMD_VAL_ADDRESSS=%s", o.wasmChain.address),
			},
			PortBindings: map[docker.Port][]docker.PortBinding{},
			Mounts:       []string{fmt.Sprintf("%s:/config", configDir)},
			Entrypoint: []string{
				"sh",
				"-c",
				"chmod +x config/wasm_bootstrap.sh && config/wasm_bootstrap.sh",
			},
		},
		noRestart,
	)
	if err != nil {
		return err
	}

	err = o.setTendermintEndpoint()
	if err != nil {
		return err
	}

	return o.setGrpcEndpoint()
}

func (o *Orchestrator) initWasmConfigs() (dir string, err error) {
	dir, err = os.MkdirTemp("", "e2e-configs")
	if err != nil {
		return
	}

	if err != nil {
		fmt.Println(err)
	}

	err = copyFiles("./config/", dir)
	if err != nil {
		return
	}

	configPath := filepath.Join(dir, "config.toml")
	config := tmconfig.DefaultConfig()
	config.P2P.ListenAddress = "tcp://0.0.0.0:26656"
	config.RPC.ListenAddress = fmt.Sprintf("tcp://0.0.0.0:%s", WASMD_TRPC_PORT)
	config.StateSync.Enable = false
	config.P2P.AddrBookStrict = false
	config.P2P.Seeds = ""
	tmconfig.WriteConfigFile(configPath, config)

	return
}

func (o *Orchestrator) wasmdBlockHeight() (int64, error) {
	status, err := o.wasmRPC.Status(context.Background())
	if err != nil {
		return 0, err
	}

	return status.SyncInfo.LatestBlockHeight, nil
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

func (o *Orchestrator) setTendermintEndpoint() (err error) {
	path := o.wasmdResource.GetHostPort(fmt.Sprintf("%s/tcp", WASMD_TRPC_PORT))
	endpoint := fmt.Sprintf("tcp://%s", path)
	o.wasmRPC, err = rpchttp.New(endpoint, "/websocket")
	return
}

func (o *Orchestrator) setGrpcEndpoint() (err error) {
	o.QueryRpc = o.wasmdResource.GetHostPort(fmt.Sprintf("%s/tcp", WASM_GRPC_PORT))
	return
}

func (o *Orchestrator) setContractAddress() error {
	grpcConn, err := grpc.Dial(
		o.QueryRpc,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	defer grpcConn.Close()

	queryClient := wasmtypes.NewQueryClient(grpcConn)
	msg := wasmtypes.QueryContractsByCodeRequest{CodeId: 1}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := queryClient.ContractsByCode(ctx, &msg)
	if err != nil {
		return err
	}

	if len(resp.Contracts[0]) == 0 {
		return fmt.Errorf("contract not found")
	}

	o.ContractAddress = resp.Contracts[0]
	return nil
}

func (o *Orchestrator) deployAndInitContract() error {
	return o.execWasmCmd(
		[]string{
			"sh", "-c", "chmod +x config/contract_bootstrap.sh && config/contract_bootstrap.sh",
		})
}

func (o *Orchestrator) addRelayerToContract(contractAddress, valAddress string) error {
	addMsg := fmt.Sprintf("{\"add_relayers\":{\"relayers\":[\"%s\"]}}", valAddress)
	msg := []string{
		"wasmd", "tx", "wasm", "execute", contractAddress, addMsg,
		"--from=val", "-b=block", "--gas-prices=0.25stake", "--keyring-backend=test", "--gas=auto", "--gas-adjustment=1.3", "-y",
		fmt.Sprintf("--chain-id=%s", o.wasmChain.chainId),
		fmt.Sprintf("--home=/data/%s", o.wasmChain.chainId),
	}

	return o.execWasmCmd(msg)
}

func (o *Orchestrator) startRelayer(contractAddress string) error {
	return o.execWasmCmd(
		[]string{
			"sh", "-c", fmt.Sprintf("chmod +x config/relayer_bootstrap.sh && config/relayer_bootstrap.sh %s", contractAddress),
		})
}

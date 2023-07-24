package main

import (
	_ "embed"
	"encoding/base64"
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/umee-network/umee-infra/infra/pulumi/common/resources"

	"contracts-relayer/unit"
)

const (
	relayerpath = "/home/ubuntu"
	localpath   = "/usr/local/bin"
	fileMode    = pulumi.String("0644")
	folderMode  = pulumi.String("0755")
	port        = pulumi.Float64(22)
	user        = pulumi.String("ubuntu")
)

var (
	relayerInstallPath = fmt.Sprintf("%s/%s", relayerpath, "cw-relayer")
	relayerConfigPath  = fmt.Sprintf("%s/relayer-relayer-config.toml", relayerpath)
	contractPath       = pulumi.Sprintf("%s/%s", relayerpath, "cosmwasm-artifacts.tar.gz")
)

func (network Network) Provision(ctx *pulumi.Context, secrets []NodeSecretConfig) error {
	conf := config.New(ctx, "")
	sshPrivate := conf.RequireSecret("sshprivate").ApplyT(func(b64private string) (string, error) {
		privatebytes, err := base64.StdEncoding.DecodeString(b64private)
		if err != nil {
			return "", err
		}

		return string(privatebytes), nil
	}).(pulumi.StringOutput)

	instance, err := compute.LookupInstance(ctx, &compute.LookupInstanceArgs{
		Name: &network.NodeConfig.InstanceName,
		Zone: pulumi.StringRef(network.NodeConfig.Location.Zone),
	})
	if err != nil {
		return err
	}

	conn := remote.ConnectionArgs{
		Host:       pulumi.String(instance.NetworkInterfaces[0].AccessConfigs[0].NatIp),
		Port:       port,
		User:       user,
		PrivateKey: sshPrivate,
	}

	// reset and cleanup files
	cleanupScript := pulumi.String(cleanupServices())
	cleanup, err := remote.NewCommand(
		ctx,
		"wasm-cleanup",
		&remote.CommandArgs{
			Connection: conn,
			Create:     cleanupScript,
		},
	)
	if err != nil {
		return err
	}

	//reinit wasmd chain
	reinitChainScript := pulumi.String(reInitChain())
	reinitChain, err := remote.NewCommand(
		ctx,
		"wasm-chain-reinit",
		&remote.CommandArgs{
			Connection: conn,
			Create:     reinitChainScript,
		},
		pulumi.DependsOn([]pulumi.Resource{cleanup}),
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "10m"}),
	)
	if err != nil {
		return err
	}

	// restart wasmd chain
	restartChain, err := remote.NewCommand(
		ctx,
		"wasm-chain-restart",
		&remote.CommandArgs{
			Connection: conn,
			Create:     serviceRestartScript("wasmd"),
		},
		pulumi.DependsOn([]pulumi.Resource{cleanup, reinitChain}),
	)
	if err != nil {
		return err
	}

	//restart caddy
	restartCaddy, err := remote.NewCommand(
		ctx,
		"caddy-restart",
		&remote.CommandArgs{
			Connection: conn,
			Update:     serviceRestartScript("caddy"),
		},
		pulumi.DependsOn([]pulumi.Resource{cleanup, restartChain, reinitChain}),
	)

	// ".cw-relayer
	techName := "cw-relayer"
	relayerSpec := unit.UnitSpec{
		Name:              techName,
		Description:       fmt.Sprintf("%s daemon", techName),
		User:              "ubuntu",
		BinaryInstallPath: relayerInstallPath,
	}
	relayerUnit := relayerSpec.ToUnit(relayerConfigPath)

	// set environment for relayer keyring pass
	keyPass := conf.RequireSecret("keypass")
	environment := map[string]pulumi.StringInput{
		"CW_RELAYER_PASS": keyPass,
	}
	relayerUnit.Environment = environment

	uploadCwRelayerBinary, err := remote.NewCopyFile(ctx, relayerUnit.Name+"-"+"binary-upload", &remote.CopyFileArgs{
		Connection: conn,
		// TODO: don't assume /usr/local/ as the base path (brittle); will work for now since we control action file, may not work on a particular devs machine
		LocalPath:  pulumi.Sprintf("%s/%s", localpath, network.LocalRelayerBinary),
		RemotePath: pulumi.String(relayerInstallPath),
	}, pulumi.DependsOn([]pulumi.Resource{restartCaddy, restartChain}), pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "20m"}))
	if err != nil {
		return err
	}

	uploadContract, err := remote.NewCopyFile(ctx, relayerUnit.Name+"-"+"contract-upload", &remote.CopyFileArgs{
		Connection: conn,
		// TODO: don't assume /usr/local/ as the base path (brittle); will work for now since we control action file, may not work on a particular devs machine
		LocalPath:  pulumi.Sprintf("%s/%s", localpath, network.LocalContractTar),
		RemotePath: contractPath,
	}, pulumi.DependsOn([]pulumi.Resource{restartChain, restartCaddy, uploadCwRelayerBinary}), pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "20m"}))
	if err != nil {
		return err
	}

	// prep contracts and keyring
	prepScript := pulumi.String(prepArtifactAndKeyring())
	prep, err := remote.NewCommand(
		ctx,
		relayerUnit.Name+"-"+"prep-contracts-keyring",
		&remote.CommandArgs{
			Connection: conn,
			Create:     prepScript,
		},
		pulumi.DependsOn([]pulumi.Resource{uploadContract, restartChain}),
	)
	if err != nil {
		return err
	}

	// deploy contract
	storeContractScript := pulumi.String(deployContract())
	storeAndInitContract, err := remote.NewCommand(
		ctx,
		relayerUnit.Name+"-"+"deploy-contract",
		&remote.CommandArgs{
			Connection: conn,
			Create:     storeContractScript,
		},
		pulumi.DependsOn([]pulumi.Resource{uploadContract, prep, restartChain}),
	)
	if err != nil {
		return err
	}

	relayerConfig := unit.RelayerConfig{
		UserAddress:     network.UserAddress,
		ContractAddress: network.ContractAddress,
	}
	configBody := relayerConfig.GenRelayerConfig()
	configPath := pulumi.String(relayerConfigPath)
	configInit, err := resources.NewStringToRemoteFileCommand(ctx, relayerUnit.Name+"-"+"relayer-config", resources.StringToRemoteFileCommandArgs{
		Connection:      conn,
		Body:            configBody,
		DestinationPath: configPath,
		FileMode:        fileMode,
		FileUser:        relayerUnit.User,
		FileGroup:       relayerUnit.User,
		FolderMode:      folderMode,
		FolderUser:      relayerUnit.User,
		FolderGroup:     relayerUnit.User,
		Triggers:        pulumi.Array{configPath, configBody},
	}, pulumi.DependsOn([]pulumi.Resource{uploadCwRelayerBinary, prep, storeAndInitContract}))
	if err != nil {
		return err
	}

	// start relayer daemon, can also be removed as it already exists
	unitBody := relayerUnit.GenSystemdUnit()
	unitPath := pulumi.Sprintf("/etc/systemd/system/%s.service", relayerUnit.Name)
	_, err = resources.NewStringToRemoteFileCommand(ctx, relayerUnit.Name+"-"+"systemd-unit", resources.StringToRemoteFileCommandArgs{
		Connection:      conn,
		Body:            unitBody,
		DestinationPath: unitPath,
		FileMode:        fileMode,
		FileUser:        relayerUnit.User,
		FileGroup:       relayerUnit.User,
		FolderMode:      folderMode,
		FolderUser:      relayerUnit.User,
		FolderGroup:     relayerUnit.User,
		RunAfter:        serviceRestartScript(relayerUnit.Name),
		Triggers:        pulumi.Array{unitPath, unitBody},
	}, pulumi.DependsOn([]pulumi.Resource{configInit, storeAndInitContract}))
	if err != nil {
		return err
	}

	return nil
}

func serviceRestartScript(service string) pulumi.StringOutput {
	return pulumi.Sprintf(`sudo systemctl daemon-reload && sudo systemctl enable %s && sudo systemctl start %s`, service, service)
}

package main

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"path"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/umee-network/umee-infra/infra/pulumi/common/resources"

	"contracts-relayer/unit"
)

const (
	relayerpath = "/home/ubuntu/"
	localpath   = "/usr/local/bin"
)

func (network Network) Provision(ctx *pulumi.Context, secrets []NodeSecretConfig) error {
	var addrs pulumi.StringArray

	conf := config.New(ctx, "")
	sshPrivate := conf.RequireSecret("sshprivate").ApplyT(func(b64private string) (string, error) {
		privatebytes, err := base64.StdEncoding.DecodeString(b64private)
		if err != nil {
			return "", err
		}

		return string(privatebytes), nil
	}).(pulumi.StringOutput)

	instanceID := pulumi.IDInput(pulumi.ID(network.NodeConfig.InstanceID))
	instance, err := compute.GetInstance(ctx, network.NodeConfig.InstanceName, instanceID, &compute.InstanceState{})

	conn := remote.ConnectionArgs{
		Host:       addrs.ToStringArrayOutput().Index(pulumi.Int(0)),
		Port:       pulumi.Float64(22),
		User:       pulumi.String("ubuntu"),
		PrivateKey: sshPrivate,
	}

	// reinit wasmd chain
	reinitChainScript := pulumi.String(reInitChain())
	reinitChain, err := remote.NewCommand(
		ctx,
		"wasm-chain-reinit",
		&remote.CommandArgs{
			Connection: conn,
			Create:     reinitChainScript,
		},
		pulumi.DependsOn([]pulumi.Resource{instance}),
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
			Create: pulumi.Sprintf(`
						    set -e
							sudo systemctl restart wasmd.service`,
			),
		}, pulumi.DependsOn([]pulumi.Resource{reinitChain}),
	)
	if err != nil {
		return err
	}

	// restart caddy
	restartCaddy, err := remote.NewCommand(
		ctx,
		"wasm-chain-restart",
		&remote.CommandArgs{
			Connection: conn,
			Create: pulumi.Sprintf(`
						    set -e
							sudo systemctl restart caddy.service`,
			),
		}, pulumi.DependsOn([]pulumi.Resource{reinitChain}),
	)

	// ".cw-relayer"
	techName := "cw-relayer"
	relayerSpec := unit.UnitSpec{
		Name:              techName,
		Description:       fmt.Sprintf("%s daemon", techName),
		User:              "ubuntu",
		BinaryInstallPath: fmt.Sprintf("%s/%s", localpath, network.LocalRelayerBinary),
	}
	relayerUnit := relayerSpec.ToUnit(fmt.Sprintf("%s/relayer-config.toml", relayerpath))

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
		RemotePath: pulumi.Sprintf("%s/%s", relayerpath, network.LocalRelayerBinary),
	}, pulumi.DependsOn([]pulumi.Resource{restartChain, instance}))
	if err != nil {
		return err
	}

	uploadContract, err := remote.NewCopyFile(ctx, relayerUnit.Name+"-"+"contract-upload", &remote.CopyFileArgs{
		Connection: conn,
		// TODO: don't assume /usr/local/ as the base path (brittle); will work for now since we control action file, may not work on a particular devs machine
		LocalPath:  pulumi.Sprintf("%s/%s", localpath, network.LocalContractTar),
		RemotePath: pulumi.Sprintf("%s/%s", relayerpath, network.LocalContractTar),
	}, pulumi.DependsOn([]pulumi.Resource{uploadCwRelayerBinary, instance}))
	if err != nil {
		return err
	}

	// /home/ubuntu/tarfile
	unzipContracts, err := remote.NewCommand(
		ctx,
		relayerUnit.Name+"-"+"contract-unzip",
		&remote.CommandArgs{
			Connection: conn,
			Create: pulumi.Sprintf(`
						    set -e
							tar -zxvf /home/ubuntu/%s /home/ubuntu/
							cp /home/ubuntu/std_reference.wasm /home/ubuntu/
							rm -r /home/ubuntu/%s
							rm -r /home/ubuntu/cosmwasm
						`, network.LocalContractTar, network.LocalContractTar),
		}, pulumi.DependsOn([]pulumi.Resource{uploadContract}),
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
		pulumi.DependsOn([]pulumi.Resource{unzipContracts, restartChain}),
	)
	if err != nil {
		return err
	}

	installCwRelayerBinary, err := remote.NewCommand(
		ctx,
		relayerUnit.Name+"-"+"binary-install",
		&remote.CommandArgs{
			Connection: conn,
			Create: pulumi.Sprintf(`
						    set -e
							sudo cp /home/ubuntu/%s /usr/local/bin/
							sudo chmod a+x /usr/local/bin/%s
						`, network.LocalRelayerBinary, network.LocalRelayerBinary),
		}, pulumi.DependsOn([]pulumi.Resource{storeAndInitContract, uploadCwRelayerBinary}),
	)
	if err != nil {
		return err
	}

	relayerConfig := unit.RelayerConfig{
		UserAddress:     network.UserAddress,
		ContractAddress: network.ContractAddress,
	}
	configBody := relayerConfig.GenRelayerConfig()
	configPath := pulumi.Sprintf("%/relayer-config.toml", relayerpath)
	configInit, err := resources.NewStringToRemoteFileCommand(ctx, relayerUnit.Name+"-"+"relayer-config", resources.StringToRemoteFileCommandArgs{
		Connection:      conn,
		Body:            configBody,
		DestinationPath: configPath,
		FileMode:        pulumi.String("0644"),
		FileUser:        relayerUnit.User,
		FileGroup:       relayerUnit.User,
		FolderMode:      pulumi.String("0755"),
		FolderUser:      relayerUnit.User,
		FolderGroup:     relayerUnit.User,
		Triggers:        pulumi.Array{configPath, configBody},
	}, pulumi.DependsOn([]pulumi.Resource{instance, installCwRelayerBinary}))
	if err != nil {
		return err
	}

	// start relayer demon, can also be removed as it already exists
	unitBody := relayerUnit.GenSystemdUnit()
	unitPath := pulumi.String(path.Join("/etc/systemd/system", relayerUnit.Name+".service"))
	relayerInstall, err := resources.NewStringToRemoteFileCommand(ctx, relayerUnit.Name+"-"+"systemd-unit", resources.StringToRemoteFileCommandArgs{
		Connection:      conn,
		Body:            unitBody,
		DestinationPath: unitPath,
		FileMode:        pulumi.String("0644"),
		FileUser:        relayerUnit.User,
		FileGroup:       relayerUnit.User,
		FolderMode:      pulumi.String("0755"),
		FolderUser:      relayerUnit.User,
		FolderGroup:     relayerUnit.User,
		RunAfter:        pulumi.Sprintf("sudo systemctl daemon-reload && sudo systemctl enable %s", relayerUnit.Name),
		Triggers:        pulumi.Array{unitPath, unitBody},
	}, pulumi.DependsOn([]pulumi.Resource{instance, configInit}))
	if err != nil {
		return err
	}

	rebootDeps := []pulumi.Resource{
		installCwRelayerBinary,
		configInit,
		relayerInstall,
		restartCaddy,
	}

	_, err = remote.NewCommand(
		ctx,
		relayerUnit.Name+"-"+"relayer-reboot",
		&remote.CommandArgs{
			Connection: conn,
			Update:     pulumi.String("echo updates disabled..."),
			Create:     pulumi.String("sleep 30 && sudo shutdown -r 1"),
		},
		pulumi.DependsOn(rebootDeps),
	)
	if err != nil {
		return err
	}

	return nil
}

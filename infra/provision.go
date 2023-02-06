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

func (network Network) Provision(ctx *pulumi.Context, secrets []NodeSecretConfig) error {
	var addrs pulumi.StringArray
	var nodeHostNames []string

	moniker := genMoniker(network.ChainID)
	
	conf := config.New(ctx, "")
	// TODO
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

	// ".cw-relayer"
	techName := network.RelayerHomeFolderName[1:]
	relayerSpec := unit.UnitSpec{
		Name:              techName,
		Description:       fmt.Sprintf("%s daemon", techName),
		User:              "blockchain",
		BinaryInstallPath: fmt.Sprintf("/usr/local/bin/%s", network.LocalRelayerBinary),
	}
	relayerUnit := relayerSpec.ToUnit(fmt.Sprintf("/home/ubuntu/%s/relayer-config.toml", network.RelayerHomeFolderName))

	// set environment for relayer keyring pass
	// write node secrets here
	environment := map[string]string{
		"CW_RELAYER_PASS": "PASS",
	}
	relayerUnit.Environment = pulumi.ToStringMap(environment)

	uploadCwRelayerBinary, err := remote.NewCopyFile(ctx, "relayer"+relayerUnit.Name+"-cp-cosmos-binary", &remote.CopyFileArgs{
		Connection: conn,
		// TODO: don't assume /usr/local/ as the base path (brittle); will work for now since we control action file, may not work on a particular devs machine
		LocalPath:  pulumi.Sprintf("/usr/local/bin/%s", network.LocalRelayerBinary),
		RemotePath: pulumi.Sprintf("/home/ubuntu/%s", network.LocalRelayerBinary),
	}, pulumi.DependsOn([]pulumi.Resource{instance}))
	if err != nil {
		return err
	}

	installCwRelayerBinary, err := remote.NewCommand(
		ctx,
		moniker+"-"+relayerUnit.Name+"wasm-install-cosmos-binary",
		&remote.CommandArgs{
			Connection: conn,
			Create: pulumi.Sprintf(`
						    set -e
							sudo cp /home/ubuntu/%s /usr/local/bin/
							sudo chmod a+x /usr/local/bin/%s
						`, network.LocalRelayerBinary, network.LocalRelayerBinary),
		}, pulumi.DependsOn([]pulumi.Resource{uploadCwRelayerBinary}),
	)
	if err != nil {
		return err
	}

	relayerConfig := unit.RelayerConfig{
		UserAddress:     network.UserAddress,
		ContractAddress: network.ContractAddress,
	}
	configBody := relayerConfig.GenRelayerConfig()
	configPath := pulumi.String(fmt.Sprintf("/home/ubuntu/%s/relayer-config.toml", network.RelayerHomeFolderName))
	configInit, err := resources.NewStringToRemoteFileCommand(ctx, moniker+"-"+relayerUnit.Name+"-relayer-config", resources.StringToRemoteFileCommandArgs{
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
	relayerInstall, err := resources.NewStringToRemoteFileCommand(ctx, moniker+"-"+relayerUnit.Name+"-systemd-unit", resources.StringToRemoteFileCommandArgs{
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
	}, pulumi.DependsOn([]pulumi.Resource{instance, installCwRelayerBinary, configInit}))
	if err != nil {
		return err
	}

	rebootDeps := []pulumi.Resource{
		installCwRelayerBinary,
		configInit,
		relayerInstall,
	}

	_, err = remote.NewCommand(
		ctx,
		moniker+"-reboot",
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

	ctx.Export("node-hostnames", pulumi.ToStringArray(nodeHostNames))

	return nil
}

func genMoniker(chainID string) string {
	return fmt.Sprintf("cosmwasm-devnet-n%v", chainID)
}

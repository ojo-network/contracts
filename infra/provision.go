package main

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/umee-network/umee-infra/infra/pulumi/common/resources"

	"contracts/unit"
)

func (network Network) Provision(ctx *pulumi.Context, secrets []NodeSecretConfig) error {
	var addrs pulumi.StringArray
	var nodeHostNames []string

	moniker := genMoniker(network.ChainID)
	location := network.NodeConfig.Location
	addr, err := compute.NewAddress(ctx, moniker+"-ip", &compute.AddressArgs{
		Labels: pulumi.StringMap{
			"chain_id": pulumi.String(network.ChainID),
		},
		NetworkTier: pulumi.String("STANDARD"),
		Region:      pulumi.String(location.Region),
	})
	if err != nil {
		return err
	}

	addrs = append(addrs, addr.Address)

	nodeHostName := fmt.Sprintf("%s.%s.node.ojo.network", moniker, network.ChainID)
	nodeHostNames = append(nodeHostNames, nodeHostName)

	bootDisk := &compute.InstanceBootDiskArgs{
		DeviceName: pulumi.String(fmt.Sprintf("%s-bootdisk", moniker)),
		InitializeParams: &compute.InstanceBootDiskInitializeParamsArgs{
			Image: pulumi.String("family/ubuntu-minimal-2204-lts"),
			Type:  pulumi.String(network.NodeConfig.DiskType),
			Size:  pulumi.Int(network.NodeConfig.DiskSizeGB),
		},
	}

	conf := config.New(ctx, "")
	sshPublic := conf.Require("sshpublic")
	// TODO
	sshPrivate := conf.RequireSecret("sshprivate").ApplyT(func(b64private string) (string, error) {
		privatebytes, err := base64.StdEncoding.DecodeString(b64private)
		if err != nil {
			return "", err
		}

		return string(privatebytes), nil
	}).(pulumi.StringOutput)

	ubuntuPubkey := pulumi.String("ubuntu:" + sshPublic)

	serviceAccount, err := createServiceAccount(ctx, moniker+"-svc", "service account for "+moniker)
	if err != nil {
		return err
	}

	startupScript := pulumi.String(genStartupScript())
	instance, err := compute.NewInstance(ctx, moniker+"-instance", &compute.InstanceArgs{
		Name: pulumi.String(moniker),
		Labels: pulumi.StringMap{
			"chain_id": pulumi.String(network.ChainID),
		},
		MachineType:            pulumi.String(network.NodeConfig.MachineType),
		Zone:                   pulumi.String(location.Zone),
		Hostname:               pulumi.String(nodeHostNames[0]),
		AllowStoppingForUpdate: pulumi.Bool(true),
		BootDisk:               bootDisk,
		MetadataStartupScript:  startupScript,
		Metadata: pulumi.StringMap{
			"ssh-keys":               ubuntuPubkey,
			"block-project-ssh-keys": pulumi.String("true"),
		},
		ServiceAccount: &compute.InstanceServiceAccountArgs{
			Email: serviceAccount.Email,
			Scopes: pulumi.StringArray{
				pulumi.String("cloud-platform"),
			},
		},
		NetworkInterfaces: compute.InstanceNetworkInterfaceArray{
			&compute.InstanceNetworkInterfaceArgs{
				Network: pulumi.String("default"),
				AccessConfigs: compute.InstanceNetworkInterfaceAccessConfigArray{
					compute.InstanceNetworkInterfaceAccessConfigArgs{
						NatIp:       addrs[0],
						NetworkTier: pulumi.String("STANDARD"),
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	conn := remote.ConnectionArgs{
		Host:       addrs.ToStringArrayOutput().Index(pulumi.Int(0)),
		Port:       pulumi.Float64(22),
		User:       pulumi.String("ubuntu"),
		PrivateKey: sshPrivate,
	}

	startupScriptsComplete, err := remote.NewCommand(
		ctx,
		moniker+"-bootstrap-script-wait-until-ready",
		&remote.CommandArgs{
			Triggers:   pulumi.Array{startupScript},
			Connection: conn,
			Update:     pulumi.String("echo updates disabled..."),
			Create: pulumi.Sprintf(`
	             for VARIABLE in 1 2 3 4 5 6 7 8 9 .. N
	             do
	               if test -f "/tmp/STARTUP_FINISHED"; then
	                 exit 0
	               else
	                 echo 'System startup script incomplete; sleeping 30 seconds...'
	                 sleep 45
	               fi
	             done
	
	             echo 'Machine is not ready or system startup script did not complete (timeout)'
	             exit 1
	           `),
		},
		pulumi.DependsOn([]pulumi.Resource{instance}),
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "10m"}),
	)
	if err != nil {
		return err
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
	environment := map[string]string{
		"CW_RELAYER_PASS": "PASS",
	}
	relayerUnit.Environment = pulumi.ToStringMap(environment)

	uploadCwRelayerBinary, err := remote.NewCopyFile(ctx, "relayer"+relayerUnit.Name+"-cp-cosmos-binary", &remote.CopyFileArgs{
		Connection: conn,
		// TODO: don't assume /usr/local/ as the base path (brittle); will work for now since we control action file, may not work on a particular devs machine
		LocalPath:  pulumi.Sprintf("/usr/local/bin/%s", network.LocalRelayerBinary),
		RemotePath: pulumi.Sprintf("/home/ubuntu/%s", network.LocalRelayerBinary),
	}, pulumi.DependsOn([]pulumi.Resource{startupScriptsComplete, instance}))
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

	// might have to do changes here based on address and keyring
	filePath, err := filepath.Abs("./unit/config.toml")
	if err != nil {
		return err
	}

	relayerConfig, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	configBody := pulumi.String(relayerConfig)
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

	// start relayer demon
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

func createServiceAccount(ctx *pulumi.Context, name string, desc string) (*serviceaccount.Account, error) {
	account, err := serviceaccount.NewAccount(ctx, name, &serviceaccount.AccountArgs{
		AccountId:   pulumi.String(name),
		DisplayName: pulumi.String(name),
		Description: pulumi.String(desc),
	})
	if err != nil {
		return nil, err
	}

	iamMember := account.Email.ApplyT(func(email string) string {
		return "serviceAccount:" + email
	}).(pulumi.StringOutput)

	gcpProject, ok := ctx.GetConfig("gcp:project")
	if !ok {
		return nil, fmt.Errorf("gcp:project must be set")
	}

	_, err = projects.NewIAMMember(ctx, name+"-metricwriter-role", &projects.IAMMemberArgs{
		Role:    pulumi.String("roles/monitoring.metricWriter"),
		Member:  iamMember,
		Project: pulumi.String(gcpProject),
	})
	if err != nil {
		return nil, err
	}

	_, err = projects.NewIAMMember(ctx, name+"-logwriter-role", &projects.IAMMemberArgs{
		Role:    pulumi.String("roles/logging.logWriter"),
		Member:  iamMember,
		Project: pulumi.String(gcpProject),
	})
	if err != nil {
		return nil, err
	}

	return account, nil
}

func genMoniker(chainID string) string {
	return fmt.Sprintf("cosmwasm-devnet-n%v", chainID)
}

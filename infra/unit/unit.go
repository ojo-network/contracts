package unit

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type UnitSpec struct {
	Name              string
	Description       string
	User              string
	BinaryInstallPath string
}

func (u UnitSpec) ToUnit(cmdStr string) Unit {
	cmdParts := []string{u.BinaryInstallPath, cmdStr}
	cmd := strings.Join(cmdParts, " ")

	return Unit{
		Name:        u.Name,
		Description: u.Description,
		User:        pulumi.String(u.User),
		ExecStart:   pulumi.String(cmd),
	}
}

type Unit struct {
	Name        string
	Description string
	User        pulumi.StringInput
	Environment pulumi.StringMap
	ExecStart   pulumi.StringInput
}

func (u Unit) GenSystemdUnit() pulumi.StringOutput {
	type templateArgs struct {
		Name        string
		Description string
		User        string
		Environment map[string]string
		ExecStart   string
	}

	return pulumi.All(u.Environment, u.ExecStart, u.User).ApplyT(func(args []interface{}) (string, error) {
		environment := args[0].(map[string]string)
		execStart := args[1].(string)
		user := args[2].(string)

		var buf bytes.Buffer
		err := unitServiceTemplate.Execute(&buf, templateArgs{
			Description: u.Description,
			User:        user,
			Environment: environment,
			ExecStart:   execStart,
		})
		if err != nil {
			return "", err
		}

		return buf.String(), nil
	}).(pulumi.StringOutput)
}

type RelayerConfig struct {
	UserAddress     string
	ContractAddress string
}

func (u RelayerConfig) GenRelayerConfig() pulumi.StringOutput {
	return pulumi.All(u.UserAddress, u.ContractAddress).ApplyT(func(args []interface{}) (string, error) {
		userAddress := args[0].(string)
		contractAddress := args[1].(string)

		var buf bytes.Buffer
		err := relayerConfigTemplate.Execute(&buf, RelayerConfig{
			ContractAddress: contractAddress,
			UserAddress:     userAddress,
		})
		if err != nil {
			return "", err
		}

		return buf.String(), nil
	}).(pulumi.StringOutput)
}

var (
	//go:embed unit.service.tmpl
	unitServiceTemplateStr string
	unitServiceTemplate    *template.Template

	//go:embed config.toml.tmpl
	relayerConfigStr      string
	relayerConfigTemplate *template.Template
)

func init() {
	var err error
	unitServiceTemplate, err = template.New("unitServiceTemplate").Parse(unitServiceTemplateStr)
	if err != nil {
		log.Fatal(err)
	}

	relayerConfigTemplate, err = template.New("relayerConfigTemplate").Parse(relayerConfigStr)
	if err != nil {
		log.Fatal(err)
	}
}

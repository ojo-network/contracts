package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

const (
	defaultListenAddr      = "0.0.0.0:7171"
	defaultSrvWriteTimeout = 15 * time.Second
	defaultSrvReadTimeout  = 15 * time.Second
	defaultProviderTimeout = 100 * time.Millisecond
)

var (
	validate = validator.New()

	// ErrEmptyConfigPath defines a sentinel error for an empty config path.
	ErrEmptyConfigPath = errors.New("empty configuration file path")
)

type (
	// Config defines all necessary cw-relayer configuration parameters.
	Config struct {
		Account         Account `mapstructure:"account" validate:"required,gt=0,dive,required"`
		Keyring         Keyring `mapstructure:"keyring" validate:"required,gt=0,dive,required"`
		RPC             RPC     `mapstructure:"rpc" validate:"required,gt=0,dive,required"`
		GasAdjustment   float64 `mapstructure:"gas_adjustment" validate:"required"`
		ContractAddress string  `mapstructure:"contract_address" validate:"required"`
		ProviderTimeout string  `mapstructure:"provider_timeout"`
		QueryRPC        string  `mapstructure:"query_rpc"`
	}

	// Account defines account related configuration that is related to the Client
	// Network and Receives Pricing information.
	Account struct {
		ChainID string `mapstructure:"chain_id" validate:"required"`
		Address string `mapstructure:"address" validate:"required"`
	}

	// Keyring defines the required Client-chain keyring configuration.
	Keyring struct {
		Backend string `mapstructure:"backend" validate:"required"`
		Dir     string `mapstructure:"dir" validate:"required"`
	}

	// RPC defines RPC configuration of both the Ojo gRPC and Tendermint nodes.
	RPC struct {
		TMRPCEndpoint string `mapstructure:"tmrpc_endpoint" validate:"required"`
		GRPCEndpoint  string `mapstructure:"grpc_endpoint" validate:"required"`
		RPCTimeout    string `mapstructure:"rpc_timeout" validate:"required"`
	}

	Relay struct {
		SymbolRates [][2]string `json:"symbol_rates,omitempty"`
		ResolveTime uint64      `json:"resolve_time,omitempty"`
		RequestID   uint64      `json:"request_id,omitempty"`
	}
)

// Validate returns an error if the Config object is invalid.
func (c Config) Validate() error {
	return validate.Struct(c)
}

// ParseConfig attempts to read and parse configuration from the given file path.
// An error is returned if reading or parsing the config fails.
func ParseConfig(configPath string) (Config, error) {
	var cfg Config

	if configPath == "" {
		return cfg, ErrEmptyConfigPath
	}

	viper.AutomaticEnv()
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("failed to read config: %w", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode config: %w", err)
	}

	return cfg, cfg.Validate()
}

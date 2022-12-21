package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

const (
	defaultProviderTimeout = 100 * time.Millisecond
	defaultQueryRPC        = "0.0.0.0:9091"
	defaultMissedThreshold = 5
	defaultTimeoutHeight   = 5
)

var (
	validate = validator.New()

	// ErrEmptyConfigPath defines a sentinel error for an empty config path.
	ErrEmptyConfigPath = errors.New("empty configuration file path")
)

type (
	// Config defines all necessary cw-relayer configuration parameters.
	Config struct {
		Account Account `mapstructure:"account" validate:"required,gt=0,dive,required"`
		Keyring Keyring `mapstructure:"keyring" validate:"required,gt=0,dive,required"`
		RPC     RPC     `mapstructure:"rpc" validate:"required,gt=0,dive,required"`

		ProviderTimeout string `mapstructure:"provider_timeout"`
		ContractAddress string `mapstructure:"contract_address"`
		TimeoutHeight   int64  `mapsturture:"timeout_height"`

		// force relay prices and reset epoch time in contracts if err in broadcasting tx
		MissedThreshold int64 `mapstructure:"missed_threshold"`

		GasAdjustment float64 `mapstructure:"gas_adjustment" validate:"required"`
		GasPrices     string  `mapstructure:"gas_prices" validate:"required"`
		Fees          string  `mapstructure:"fees" validate:"required"`
		// query rpc for ojo node
		QueryRPC string `mapstructure:"query_rpc"`
	}

	// Account defines account related configuration that is related to the Client
	// Network and Receives Pricing information.
	Account struct {
		AccPrefix string `mapstructure:"acc_prefix" validate:"required"`
		ChainID   string `mapstructure:"chain_id" validate:"required"`
		Address   string `mapstructure:"address" validate:"required"`
	}

	// Keyring defines the required Client-chain keyring configuration.
	Keyring struct {
		Backend string `mapstructure:"backend" validate:"required"`
		Dir     string `mapstructure:"dir" validate:"required"`
	}

	// RPC defines RPC configuration of both the wasmd chain gRPC and Tendermint nodes.
	RPC struct {
		TMRPCEndpoint string `mapstructure:"tmrpc_endpoint" validate:"required"`
		GRPCEndpoint  string `mapstructure:"grpc_endpoint" validate:"required"`
		RPCTimeout    string `mapstructure:"rpc_timeout" validate:"required"`
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

	if len(cfg.ProviderTimeout) == 0 {
		cfg.ProviderTimeout = defaultProviderTimeout.String()
	}

	if len(cfg.QueryRPC) == 0 {
		cfg.QueryRPC = defaultQueryRPC
	}

	if len(cfg.ContractAddress) == 0 {
		return cfg, fmt.Errorf("contract address cannot be nil")
	}

	if cfg.MissedThreshold <= 0 {
		cfg.MissedThreshold = defaultMissedThreshold
	}

	if cfg.TimeoutHeight == 0 {
		cfg.TimeoutHeight = defaultTimeoutHeight
	}

	return cfg, cfg.Validate()
}

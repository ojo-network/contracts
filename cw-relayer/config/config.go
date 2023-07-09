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
	defaultTimeoutHeight   = 5
	defaultTimeout         = 1 * time.Minute
	defaultRetries         = 1
	defaultTickEventType   = "ojo.oracle.v1.EventSetFxRate"
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

		TargetRPC RPC     `mapstructure:"target_rpc" validate:"required,gt=0,dive,required"`
		DataRPC   DataRpc `mapstructure:"data_rpc" validate:"required,gt=0,dive,required"`

		Timeout Timeout `mapstructure:"timeout" validate:"required,gt=0,dive,required"`
		Gas     Gas     `mapstructure:"gas" validate:"required,gt=0,dive,required"`

		MaxRetries   int64  `mapstructure:"max_retries" validate:"required"`
		PingDuration string `mapstructure:"ping_duration" validate:"required"`
		NumBundle    int64  `mapstructure:"num_bundle"`

		ContractAddress string `mapstructure:"contract_address"`

		// if true, would ignore any errors when querying median or deviations
		IgnoreMedianErrors bool `mapstructure:"ignore_median_errors"`

		// query rpc for ojo node
		TickEventType string `mapstructure:"event_type"`

		BlockHeightConfig BlockHeightConfig `mapstructure:"block_height_config" validate:"required,dive,required"`
	}

	Timeout struct {
		EventTimeout    string `mapstructure:"event_timeout" validate:"required"`
		MaxTickTimeout  string `mapstructure:"max_tick_timeout" validate:"required"`
		QueryTimeout    string `mapstructure:"query_timeout" validate:"required"`
		TimeoutHeight   int64  `mapstructure:"timeout_height" validate:"required"`
		ProviderTimeout string `mapstructure:"provider_timeout" validate:"required"`
	}

	Gas struct {
		GasAdjustment float64 `mapstructure:"gas_adjustment" validate:"required"`
		GasPrices     string  `mapstructure:"gas_prices" validate:"required"`
		GasLimitPerTx float64 `mapstructure:"gas_per_tx" validate:"required"`
	}

	BlockHeightConfig struct {
		SkipError bool `mapstructure:"skip_error" validate:"required"`
	}

	DataRpc struct {
		QueryRPCS []string `mapstructure:"query_rpcs" validate:"required"`
		EventRPCS []string `mapstructure:"event_rpcs" validate:"required"`
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

	RestartConfig struct {
		AutoID    bool   `mapstructure:"auto_id"`
		Denom     string `mapstructure:"denom"`
		SkipError bool   `mapstructure:"skip_error"`
	}

	// RPC defines RPC configuration of both the wasmd chain and Tendermint nodes.
	RPC struct {
		TMRPCEndpoint []string `mapstructure:"tmrpc_endpoint" validate:"required"`
		RPCTimeout    string   `mapstructure:"rpc_timeout" validate:"required"`
		QueryEndpoint []string `mapstructure:"query_endpoint" validate:"required"`
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

	if len(cfg.Timeout.ProviderTimeout) == 0 {
		cfg.Timeout.ProviderTimeout = defaultProviderTimeout.String()
	}

	if len(cfg.DataRPC.QueryRPCS) == 0 {
		cfg.DataRPC.QueryRPCS = []string{defaultQueryRPC}
	}

	if len(cfg.ContractAddress) == 0 {
		return cfg, fmt.Errorf("contract address cannot be nil")
	}

	if cfg.Timeout.TimeoutHeight == 0 {
		cfg.Timeout.TimeoutHeight = defaultTimeoutHeight
	}

	if len(cfg.Timeout.EventTimeout) == 0 {
		cfg.Timeout.EventTimeout = defaultTimeout.String()
	}

	if len(cfg.Timeout.MaxTickTimeout) == 0 {
		cfg.Timeout.MaxTickTimeout = defaultTimeout.String()
	}

	if len(cfg.Timeout.QueryTimeout) == 0 {
		cfg.Timeout.QueryTimeout = defaultTimeout.String()
	}

	if len(cfg.TickEventType) == 0 {
		cfg.TickEventType = defaultTickEventType
	}

	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = defaultRetries
	}

	return cfg, cfg.Validate()
}

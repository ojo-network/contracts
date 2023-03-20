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
	defaultResolveDuration = 2 * time.Second
	defaultQueryRetries    = 1
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
		Account Account       `mapstructure:"account" validate:"required,gt=0,dive,required"`
		Keyring Keyring       `mapstructure:"keyring" validate:"required,gt=0,dive,required"`
		RPC     RPC           `mapstructure:"rpc" validate:"required,gt=0,dive,required"`
		Restart RestartConfig `mapstructure:"restart" validate:"required"`

		ProviderTimeout string `mapstructure:"provider_timeout"`
		ContractAddress string `mapstructure:"contract_address"`
		TimeoutHeight   int64  `mapstructure:"timeout_height"`
		EventTimeout    string `mapstructure:"event_timeout"`
		MaxTickTimeout  string `mapstructure:"max_tick_timeout"`
		QueryTimeout    string `mapstructure:"query_timeout"`
		MaxQueryRetries int64  `mapstructure:"max_query_retries"`

		MedianRequestID    uint64 `mapstructure:"median_request_id"`
		RequestID          uint64 `mapstructure:"request_id"`
		DeviationRequestID uint64 `mapstructure:"deviation_request_id"`

		// force relay prices and reset epoch time in contracts if err in broadcasting tx
		MissedThreshold int64  `mapstructure:"missed_threshold"`
		MedianDuration  int64  `mapstructure:"median_duration"`
		ResolveDuration string `mapstructure:"resolve_duration"`

		GasAdjustment float64 `mapstructure:"gas_adjustment" validate:"required"`
		GasPrices     string  `mapstructure:"gas_prices" validate:"required"`

		// query rpc for ojo node
		QueryRPCS     []string `mapstructure:"query_rpcs" validate:"required"`
		EventRPCS     []string `mapstructure:"event_rpcs" validate:"required"`
		TickEventType string   `mapstructure:"event_type"`
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
		TMRPCEndpoint string `mapstructure:"tmrpc_endpoint" validate:"required"`
		RPCTimeout    string `mapstructure:"rpc_timeout" validate:"required"`
		QueryEndpoint string `mapstructure:"query_endpoint" validate:"required"`
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

	if len(cfg.QueryRPCS) == 0 {
		cfg.QueryRPCS = []string{defaultQueryRPC}
	}

	if len(cfg.ContractAddress) == 0 {
		return cfg, fmt.Errorf("contract address cannot be nil")
	}

	if cfg.TimeoutHeight == 0 {
		cfg.TimeoutHeight = defaultTimeoutHeight
	}

	if len(cfg.EventTimeout) == 0 {
		cfg.EventTimeout = defaultTimeout.String()
	}

	if len(cfg.MaxTickTimeout) == 0 {
		cfg.MaxTickTimeout = defaultTimeout.String()
	}

	if len(cfg.QueryTimeout) == 0 {
		cfg.QueryTimeout = defaultTimeout.String()
	}

	if len(cfg.ResolveDuration) == 0 {
		cfg.ResolveDuration = defaultResolveDuration.String()
	}

	if len(cfg.TickEventType) == 0 {
		cfg.TickEventType = defaultTickEventType
	}
	if cfg.MaxQueryRetries == 0 {
		cfg.MaxQueryRetries = defaultQueryRetries
	}

	return cfg, cfg.Validate()
}

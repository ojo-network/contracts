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
	defaultProviderTimeout = 100 * time.Millisecond
	//TODO: add default ojo rpc
	defaultQueryRPC        = ""
	defaultMissedThreshold = 5
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
		ContractAddress string  `mapstructure:"contract_address"`
		ProviderTimeout string  `mapstructure:"provider_timeout"`
		MissedThreshold int64   `mapstructure:"missed_threshold"`
		// query rpc for ojo node
		QueryRPC string `mapstructure:"query_rpc"`
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

	// RPC defines RPC configuration of both the wasmd chain gRPC and Tendermint nodes.
	RPC struct {
		TMRPCEndpoint string `mapstructure:"tmrpc_endpoint" validate:"required"`
		GRPCEndpoint  string `mapstructure:"grpc_endpoint" validate:"required"`
		RPCTimeout    string `mapstructure:"rpc_timeout" validate:"required"`
	}

	MsgRelay struct {
		Relay Msg `json:"relay"`
	}

	MsgForceRelay struct {
		Relay Msg `json:"force_relay"`
	}

	Msg struct {
		SymbolRates [][2]string `json:"symbol_rates,omitempty"`
		ResolveTime int64       `json:"resolve_time,omitempty"`
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
	
	return cfg, cfg.Validate()
}

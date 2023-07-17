package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/ojo-network/cw-relayer/config"
	"github.com/ojo-network/cw-relayer/relayer"
	relayerclient "github.com/ojo-network/cw-relayer/relayer/client"
	"github.com/ojo-network/cw-relayer/relayer/client/txbundle"
)

const (
	logLevelJSON = "json"
	logLevelText = "text"

	flagLogLevel  = "log-level"
	flagLogFormat = "log-format"

	envVariablePass = "CW_RELAYER_PASS"
)

var rootCmd = &cobra.Command{
	Use:   "cw-relayer [config-file]",
	Args:  cobra.ExactArgs(1),
	Short: "cw-relayer is a side-car process for providing Wasm-enabled chains with Ojo's pricing Data",
	Long: `cw-relayer is a side-car process for providing Wasm-enabled chains with Ojo's pricing Data,
	It queries prices from ojo node and pushes it to Wasm contracts on regular intervals`,
	RunE: cwRelayerCmdHandler,
}

func init() {
	rootCmd.PersistentFlags().String(flagLogLevel, zerolog.InfoLevel.String(), "logging level")
	rootCmd.PersistentFlags().String(flagLogFormat, logLevelText, "logging format; must be either json or text")

	rootCmd.AddCommand(getVersionCmd())
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func cwRelayerCmdHandler(cmd *cobra.Command, args []string) error {
	logLvlStr, err := cmd.Flags().GetString(flagLogLevel)
	if err != nil {
		return err
	}

	logLvl, err := zerolog.ParseLevel(logLvlStr)
	if err != nil {
		return err
	}

	logFormatStr, err := cmd.Flags().GetString(flagLogFormat)
	if err != nil {
		return err
	}

	var logWriter io.Writer
	switch strings.ToLower(logFormatStr) {
	case logLevelJSON:
		logWriter = os.Stderr

	case logLevelText:
		logWriter = zerolog.ConsoleWriter{Out: os.Stderr}

	default:
		return fmt.Errorf("invalid logging format: %s", logFormatStr)
	}

	logger := zerolog.New(logWriter).Level(logLvl).With().Timestamp().Logger()

	cfg, err := config.ParseConfig(args[0])
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(cmd.Context())
	g, ctx := errgroup.WithContext(ctx)

	// listen for and trap any OS signal to gracefully shutdown and exit
	trapSignal(cancel, logger)

	rpcTimeout, err := time.ParseDuration(cfg.TargetRPC.RPCTimeout)
	if err != nil {
		return fmt.Errorf("failed to parse RPC timeout: %w", err)
	}

	eventTimeout, err := time.ParseDuration(cfg.Timeout.EventTimeout)
	if err != nil {
		return fmt.Errorf("failed to parse Event timeout: %w", err)
	}

	tickDuration, err := time.ParseDuration(cfg.TickDuration)
	if err != nil {
		return fmt.Errorf("failed to parse Tick duration: %w", err)
	}

	maxTickTimeout, err := time.ParseDuration(cfg.Timeout.MaxTickTimeout)
	if err != nil {
		return fmt.Errorf("failed to parse Event timeout: %w", err)
	}

	queryTimeout, err := time.ParseDuration(cfg.Timeout.QueryTimeout)
	if err != nil {
		return fmt.Errorf("failed to parse Query timeout: %w", err)
	}

	pingDuration, err := time.ParseDuration(cfg.PingDuration)
	if err != nil {
		return fmt.Errorf("failed to parse Ping Duration: %w", err)
	}

	maxTimeout, err := time.ParseDuration(cfg.TxConfig.MaxTimeout)
	if err != nil {
		return fmt.Errorf("failed to parse Max Timeout: %w", err)
	}

	// Gather pass via env variable || std input
	keyringPass, err := getKeyringPassword()
	if err != nil {
		return err
	}

	// client for interacting with the ojo & wasmd chain
	client, err := relayerclient.NewRelayerClient(
		ctx,
		logger,
		cfg.Account.ChainID,
		cfg.Keyring.Backend,
		cfg.Keyring.Dir,
		keyringPass,
		cfg.TargetRPC.TMRPCEndpoint,
		cfg.TargetRPC.QueryEndpoint,
		rpcTimeout,
		cfg.Account.Address,
		cfg.Account.AccPrefix,
		cfg.Gas.GasAdjustment,
		cfg.Gas.GasPrices,
	)
	if err != nil {
		return err
	}

	uptimePing := relayer.NewUptimePing(
		logger,
		cfg.Account.Address,
		cfg.ContractAddress,
		cfg.Timeout.TimeoutHeight,
		client,
	)

	logger.Info().Msg("starting ping service")
	g.Go(func() error {
		return uptimePing.StartPing(ctx, pingDuration)
	})

	// subscribe to new block heights
	tick, err := relayerclient.NewBlockHeightSubscription(
		ctx,
		logger,
		cfg.DataRPC.EventRPCS,
		eventTimeout,
		maxTickTimeout,
		cfg.TickEventType,
		cfg.BlockHeightConfig.SkipError,
		cfg.MaxRetries,
	)
	if err != nil {
		return err
	}

	contractTick, err := relayerclient.NewContractSubscribe(
		cfg.TargetRPC.TMRPCEndpoint,
		cfg.ContractAddress,
		cfg.Account.Address,
		logger,
	)
	if err != nil {
		return err
	}

	logger.Info().Msg("starting contract subscribe service")
	g.Go(func() error {
		return contractTick.Subscribe(ctx)
	})

	logger.Info().Msg("starting price msg service")
	txBundler := txbundle.NewTxBundler(
		logger,
		cfg.TxConfig.BundleSize,
		cfg.TxConfig.MaxGasLimitPerTx,
		cfg.Timeout.TimeoutHeight,
		maxTimeout,
		client,
	)

	g.Go(func() error {
		return txBundler.Bundler(ctx)
	})

	priceService := relayer.NewPriceService(
		logger,
		cfg.DataRPC.QueryRPCS,
		cfg.MaxRetries,
		queryTimeout,
		tick.Tick,
	)

	logger.Info().Msg("starting price fetch service")
	g.Go(func() error {
		return priceService.Start(ctx)
	})

	newRelayer := relayer.New(
		logger,
		contractTick,
		priceService,
		cfg.Account.Address,
		cfg.ContractAddress,
		cfg.Timeout.TimeoutHeight,
		tickDuration,
		contractTick.Out,
		txBundler.MsgChan,
	)

	g.Go(
		func() error {
			// start the process that queries the prices on Ojo & submits them on Wasmd
			return startPriceRelayer(ctx, logger, newRelayer)
		},
	)

	// Block main process until all spawned goroutines have gracefully exited and
	// signal has been captured in the main process or if an error occurs.
	return g.Wait()
}

func getKeyringPassword() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	pass := os.Getenv(envVariablePass)
	if pass == "" {
		return input.GetString("Enter keyring password", reader)
	}
	return pass, nil
}

// trapSignal will listen for any OS signal and invoke Done on the main
// WaitGroup allowing the main process to gracefully exit.
func trapSignal(cancel context.CancelFunc, logger zerolog.Logger) {
	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGINT)

	go func() {
		sig := <-sigCh
		logger.Info().Str("signal", sig.String()).Msg("caught signal; shutting down...")
		cancel()
	}()
}

func startPriceRelayer(ctx context.Context, logger zerolog.Logger, relayer *relayer.Relayer) error {
	srvErrCh := make(chan error, 1)

	go func() {
		logger.Info().Msg("starting relayer...")
		srvErrCh <- relayer.Start(ctx)
	}()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("shutting down relayer...")
			return nil

		case err := <-srvErrCh:
			logger.Err(err).Msg("error starting the relayer")
			relayer.Stop()
			return err
		}
	}
}

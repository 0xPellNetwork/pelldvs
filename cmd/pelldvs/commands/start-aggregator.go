package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	interactorcfg "github.com/0xPellNetwork/pelldvs-interactor/config"
	aggcfg "github.com/0xPellNetwork/pelldvs/aggregator/config"
	"github.com/0xPellNetwork/pelldvs/aggregator/rpc"
	cfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/libs/cli"
	"github.com/0xPellNetwork/pelldvs/utils"
)

const (
	flagRPCAddress = "address"
	flagTimeout    = "timeout"
)

var aggregatorConfigFile string

// StartAggregatorCmd defines the command to start the aggregator
var StartAggregatorCmd = &cobra.Command{
	Use:   "start-aggregator",
	Short: "Start the aggregator",
	RunE:  startAggregator,
}

func init() {
	StartAggregatorCmd.PersistentFlags().StringVar(&aggregatorConfigFile, "config", "", "config File")
}

func init() {
	StartAggregatorCmd.Flags().String(flagRPCAddress, "", "RPC server listen address")
	StartAggregatorCmd.Flags().String(flagTimeout, "", "Aggregation operation timeout")
}

// startAggregator implements the logic to start the aggregator
func startAggregator(cmd *cobra.Command, args []string) error {
	return runAggregatorService(cmd)
}

func runAggregatorService(cmd *cobra.Command) error {
	rpcAddress := viper.GetString(flagRPCAddress)
	timeout := viper.GetString(flagTimeout)

	homeDir := cmd.Flags().Lookup(cli.HomeFlag).Value.String()

	if aggregatorConfigFile == "" {
		if homeDir == "" {
			return fmt.Errorf("home directory is required")
		}
		aggregatorConfigFile = strings.TrimRight(homeDir, "/") + "/config/aggregator.json"
	}

	aggregatorConfig, err := aggcfg.LoadConfig(aggregatorConfigFile, logger)
	if err != nil {
		return fmt.Errorf("failed to load aggregator configuration: %v", err)
	}

	if rpcAddress != "" {
		aggregatorConfig.AggregatorRPCServer = rpcAddress
	}
	if timeout != "" {
		aggregatorConfig.OperatorResponseTimeout = timeout
	}

	ctx := cmd.Context()

	db, err := cfg.DefaultDBProvider(&cfg.DBContext{
		ID:     "indexer",
		Config: config,
	})
	if err != nil {
		return fmt.Errorf("failed to init db: %v", err)
	}

	interactorConfig, err := interactorcfg.LoadConfig(config.Pell.InteractorConfigPath)
	if err != nil {
		return fmt.Errorf("failed to create interactor config from file: %v", err)
	}

	dvsReader, err := utils.CreateDVSReader(ctx, interactorConfig, db, logger)
	if err != nil {
		logger.Error("Failed to create DVS reader", "error", err)
		return fmt.Errorf("failed to create DVS reader: %v", err)
	}

	rpcAggregator, err := rpc.NewAggregatorGRPCServer(ctx, aggregatorConfig, interactorConfig, config, dvsReader, logger)
	if err != nil {
		return fmt.Errorf("failed to create RPCAggregator: %v", err)
	}

	if err = rpcAggregator.Start(); err != nil {
		return fmt.Errorf("failed to start aggregator: %v", err)
	}

	logger.Info("Aggregator started", "RPC address", rpcAddress)

	// Keep the service running until an interrupt signal is received
	select {}
}

package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/utils"
	cfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/libs/cli"
)

var (
	config = cfg.DefaultConfig()
	logger = log.NewLogger(os.Stdout)
)

func init() {
	registerFlagsRootCmd(RootCmd)
	cfg.SetGlobalCmtConfig(config)
}

func registerFlagsRootCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().String("log_level", config.LogLevel, "log level")
}

// overwritePellConfigs overwrites the pell configs with the flags
func overwritePellConfigs() {
}

// ParseConfig retrieves the default environment configuration,
// sets up the PellDVS defaultRoot and ensures that the defaultRoot exists
func ParseConfig(cmd *cobra.Command) (*cfg.Config, error) {
	conf := cfg.DefaultConfig()
	err := viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}
	home := utils.GetHomeDir(cmd)

	conf.RootDir = home

	conf.SetRoot(conf.RootDir)
	cfg.EnsureRoot(conf.RootDir)
	config = conf

	if err := conf.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("error in config file: %v", err)
	}
	if warnings := conf.CheckDeprecated(); len(warnings) > 0 {
		for _, warning := range warnings {
			logger.Info("deprecated usage found in configuration file", "usage", warning)
		}
	}

	cfg.SetGlobalCmtConfig(conf)
	cfg.SetGlobalPellConfig(conf.Pell)
	overwritePellConfigs()

	return conf, nil
}

// RootCmd is the defaultRoot command for PellDVS core.
var RootCmd = &cobra.Command{
	Use:   "pelldvs",
	Short: "BFT state machine replication for applications in any programming languages",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		if cmd.Name() == VersionCmd.Name() {
			return nil
		}

		config, err = ParseConfig(cmd)
		if err != nil {
			return err
		}

		if config.LogFormat == cfg.LogFormatJSON {
			logger = log.NewLogger(os.Stdout, log.OutputJSONOption())
		}

		if viper.GetBool(cli.TraceFlag) {
			logger = log.NewTracingLogger(logger)
		}

		logger = logger.With("module", "main")
		return nil
	},
}

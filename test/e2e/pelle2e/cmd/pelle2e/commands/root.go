package commands

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/0xPellNetwork/pelldvs/libs/cli"
	"github.com/0xPellNetwork/pelldvs/libs/log"
)

var (
	logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
)

// RootCmd is the root command for squaringd server.
var RootCmd = &cobra.Command{
	Use:   "pelle2e",
	Short: "pelle2e",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		if viper.GetBool(cli.TraceFlag) {
			logger = log.NewTracingLogger(logger)
		}
		logger = logger.With("module", "main")
		return nil
	},
}

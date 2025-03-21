package dvs

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
)

var rootLogger = log.NewLogger(os.Stdout).With("cmd", "client/dvs")

func getCmdLogger(cmd *cobra.Command) log.Logger {
	return rootLogger.With("cmd", utils.GetPrettyCommandName(cmd))
}

var DvsCmd = &cobra.Command{
	Use:   "dvs",
	Short: "Manage DVS",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	_ = chainflags.RequireFromFlagPersistentForCmds(
		addPoolsCmd,
		addSupportedChainCmd,
		createGroupCmd,
		createRegistryRouterCmd,
		pauseRegistryRouterCmd,
		setEjectionCooldownCmd,
		setEjectorCmd,
		setMinimumStakeForGroupCmd,
		setOperatorSetCmd,
		syncGroupCmd,
		unPauseRegistryRouterCmd,
		updateDVSMetadataURICmd,
		updateOperatorsCmd,
	)

	// Add commands
	DvsCmd.AddCommand(createGroupCmd)
	DvsCmd.AddCommand(setOperatorSetCmd)
	DvsCmd.AddCommand(setEjectionCooldownCmd)
	DvsCmd.AddCommand(setMinimumStakeForGroupCmd)
	DvsCmd.AddCommand(setEjectorCmd)
	DvsCmd.AddCommand(updateDVSMetadataURICmd)
	DvsCmd.AddCommand(updateOperatorsCmd)
	DvsCmd.AddCommand(createRegistryRouterCmd)
	DvsCmd.AddCommand(addSupportedChainCmd)
	DvsCmd.AddCommand(queryDVSInfoCmd)
	DvsCmd.AddCommand(pauseRegistryRouterCmd)
	DvsCmd.AddCommand(unPauseRegistryRouterCmd)
	DvsCmd.AddCommand(registerChainToPellCmd)
	DvsCmd.AddCommand(syncGroupCmd)
	DvsCmd.AddCommand(addPoolsCmd)
}

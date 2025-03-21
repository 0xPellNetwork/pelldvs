package operator

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
)

var rootLogger = log.NewLogger(os.Stdout).With("cmd", "client/operator")

func getCmdLogger(cmd *cobra.Command) log.Logger {
	return rootLogger.With("cmd", utils.GetPrettyCommandName(cmd))
}

var OperatorCmd = &cobra.Command{
	Use:     "operator",
	Aliases: []string{"o", "op"},
	Short:   "Manage Operator",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	_ = chainflags.RequireFromFlagPersistentForCmds(
		deRegisterOperatorCmd,
		modifyOperatorDetailsCmd,
		registerOperatorCmd,
		registerOperatorToDVSCmd,
		updateOperatorMetadataURICmd,
		updateSocketCmd,
	)

	// Add commands
	OperatorCmd.AddCommand(updateOperatorMetadataURICmd)
	OperatorCmd.AddCommand(modifyOperatorDetailsCmd)
	OperatorCmd.AddCommand(registerOperatorToDVSCmd)
	OperatorCmd.AddCommand(deRegisterOperatorCmd)
	OperatorCmd.AddCommand(updateSocketCmd)
	OperatorCmd.AddCommand(delegationApproverCmd)
	OperatorCmd.AddCommand(isOperatorCmd)
	OperatorCmd.AddCommand(operatorDetailsCmd)
	OperatorCmd.AddCommand(registerOperatorCmd)
	OperatorCmd.AddCommand(getWeightForGroupCmd)
}

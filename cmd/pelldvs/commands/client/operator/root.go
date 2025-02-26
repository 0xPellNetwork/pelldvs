package operator

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
)

var logger = log.NewLogger(os.Stdout).With("cmd", "tx/operator")

var OperatorCmd = &cobra.Command{
	Use:     "operator",
	Aliases: []string{"o", "op"},
	Short:   "Manage Operator",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {

	// make sure the the FROM flag is valid for the following commands
	_ = chainflags.RequireFromFlagPersistentForCmds(
		updateOperatorMetadataURICmd,
		modifyOperatorDetailsCmd,
		registerOperatorToDVSCmd,
		deRegisterOperatorCmd,
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
}

package client

import (
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/dvs"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/operator"
)

var ClientCmd = &cobra.Command{
	Use:     "client",
	Aliases: []string{"c"},
	Short:   "Client operations",
	Long:    `client operations`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	chainflags.SetPellChainPersistentFlags(ClientCmd)

	ClientCmd.Flags().SortFlags = true

	ClientCmd.AddCommand(dvs.DvsCmd)
	ClientCmd.AddCommand(operator.OperatorCmd)

}

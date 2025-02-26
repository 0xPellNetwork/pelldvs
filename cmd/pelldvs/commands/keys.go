package commands

import (
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs/pkg/keys"
	"github.com/0xPellNetwork/pelldvs/pkg/utils"
)

func KeysCmd(p utils.Prompter) *cobra.Command {
	var keysCmd = &cobra.Command{
		Use:     "keys",
		Short:   "Manage the keys used in PellDVS ecosystem",
		Aliases: []string{"k"},
	}

	keysCmd.AddCommand(keys.CreateCmd(p))
	keysCmd.AddCommand(keys.ListCmd(p))
	keysCmd.AddCommand(keys.ImportCmd(p))
	keysCmd.AddCommand(keys.ExportCmd(p))
	keysCmd.AddCommand(keys.ShowCmd(p))

	return keysCmd
}

package dvs

import (
	"github.com/spf13/cobra"
)

var unPauseRegistryRouterCmd = &cobra.Command{
	Use:   "unpause-registry-router",
	Short: "unpause-registry-router",
	Long:  `unpause-registry-router`,
	Example: `

pelldvs client dvs unpause-registry-router --from pell-localnet-deployer --chain_id
pelldvs client dvs unpause-registry-router --from pell-localnet-deployer --chain_id 666

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handlePauseOrUnRegistryRouter(cmd, false)
	},
}

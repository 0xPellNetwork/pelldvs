package dvs

import (
	"github.com/spf13/cobra"
)

var pauseRegistryRouterCmd = &cobra.Command{
	Use:   "pause-registry-router",
	Short: "pause-registry-router",
	Long:  `pause-registry-router`,
	Example: `

pelldvs client dvs pause-registry-router --from pell-localnet-deployer
pelldvs client dvs pause-registry-router --from pell-localnet-deployer

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handlePauseOrUnRegistryRouter(cmd, true)
	},
}

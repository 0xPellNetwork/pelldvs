package dvs

import (
	"github.com/spf13/cobra"
)

var unPauseRegistryRouterCmd = &cobra.Command{
	Use:   "unpause-registry-router",
	Short: "unpause-registry-router",
	Long: `
pelldvs client dvs unpause-registry-router \
	--from <key-name> \
	--rpc-url <rpc-url> \
	--registry-router <registry-router>
`,
	Example: `
pelldvs client dvs unpause-registry-router \
	--from pell-localnet-deployer \
	--rpc-url http://localhost:8545 \
	--registry-router 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handlePauseOrUnRegistryRouter(cmd, false)
	},
}

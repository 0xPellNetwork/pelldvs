package chainflags

import "github.com/spf13/cobra"

func SetPellChainPersistentFlags(cmds ...*cobra.Command) {
	// Set the persistent flags for the Pell chain
	for _, cmd := range cmds {
		FromKeyNameFlag.AddToCmdPersistentFlags(cmd)
		EthRPCURLFlag.AddToCmdPersistentFlags(cmd)
		PellRegistryRouterFactoryAddress.AddToCmdPersistentFlags(cmd)
		PellDelegationManagerAddress.AddToCmdPersistentFlags(cmd)
		PellRegistryRouterAddress.AddToCmdPersistentFlags(cmd)
		PellDVSDirectoryAddress.AddToCmdPersistentFlags(cmd)
		DVSCentralSchedulerAddress.AddToCmdPersistentFlags(cmd)
	}
}

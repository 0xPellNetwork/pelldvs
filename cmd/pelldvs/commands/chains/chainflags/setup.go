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
		CentralSchedulerContractAddressFlag.AddToCmdPersistentFlags(cmd)

		//	// from key name
		//	cmd.PersistentFlags().StringVar(
		//		&FromKeyNameFlag.Value,
		//		FromKeyNameFlag.Name,
		//		"",
		//		FromKeyNameFlag.Usage,
		//	)
		//
		//	// Ethereum RPC URL
		//	cmd.PersistentFlags().StringVar(
		//		&EthRPCURLFlag.Value,
		//		EthRPCURLFlag.Name,
		//		EthRPCURLFlag.Default,
		//		EthRPCURLFlag.Usage,
		//	)
		//
		//	// Registry Router Factory contract address
		//	cmd.PersistentFlags().StringVar(
		//		&PellRegistryRouterFactoryAddress.Value,
		//		PellRegistryRouterFactoryAddress.Name,
		//		PellRegistryRouterFactoryAddress.Default,
		//		PellRegistryRouterFactoryAddress.Usage,
		//	)
		//
		//	// Delegation contract address
		//	cmd.PersistentFlags().StringVar(
		//		&PellDelegationManagerAddress.Value,
		//		PellDelegationManagerAddress.Name,
		//		PellDelegationManagerAddress.Default,
		//		PellDelegationManagerAddress.Usage,
		//	)
		//
		//	// Registry Router contract address
		//	cmd.PersistentFlags().StringVar(
		//		&PellRegistryRouterAddress.Value,
		//		PellRegistryRouterAddress.Name,
		//		PellRegistryRouterAddress.Default,
		//		PellRegistryRouterAddress.Usage,
		//	)
		//
		//	// DVS Directory contract address
		//	cmd.PersistentFlags().StringVar(
		//		&PellDVSDirectoryAddress.Value,
		//		PellDVSDirectoryAddress.Name,
		//		PellDVSDirectoryAddress.Default,
		//		PellDVSDirectoryAddress.Usage,
		//	)
		//
		//	// Registry CentralSchedulerContract contract address
		//	cmd.PersistentFlags().StringVar(
		//		&CentralSchedulerContractAddressFlag.Value,
		//		CentralSchedulerContractAddressFlag.Name,
		//		CentralSchedulerContractAddressFlag.Default,
		//		CentralSchedulerContractAddressFlag.Usage,
		//	)

	}
}

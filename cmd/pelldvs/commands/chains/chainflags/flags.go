package chainflags

var FromKeyNameFlag = &StringFlag{
	Name:  "from",
	Usage: "Name of the key to use for signing the transaction",
}

var EthRPCURLFlag = &StringFlag{
	Name:  "rpc-url",
	Usage: "URL of the Ethereum RPC server",
}

var PellRegistryRouterFactoryAddress = &StringFlag{
	Name:  "registry-router-factory",
	Usage: "Address of the registry router factory contract",
}

var PellDelegationManagerAddress = &StringFlag{
	Name:  "delegation-manager",
	Usage: "Address of the delegation contract",
}

var PellRegistryRouterAddress = &StringFlag{
	Name:  "registry-router",
	Usage: "Address of the registry router contract",
}

var PellDVSDirectoryAddress = &StringFlag{
	Name:  "dvs-directory",
	Usage: "Address of the DVS directory contract",
}

var CentralSchedulerContractAddressFlag = &StringFlag{
	Name:  "central-scheduler",
	Usage: "Address of the registry CentralScheduler contract",
}

var ChainIDFlag = &IntFlag{
	Name:  "chain-id",
	Usage: "Chain ID",
	Aliases: NewAliases(
		"chain_id",
	),
}

var DVSApproverKeyNameFlag = &StringFlag{
	Name:  "dvs-approver-key-name",
	Usage: "key name of the DVS approver",
}

var GroupNumbers = &StringFlag{
	Name:  "groups",
	Usage: "group numbers",
}

var MetadataURI = &StringFlag{
	Name:  "metadata-uri",
	Usage: "metadata URI",
}

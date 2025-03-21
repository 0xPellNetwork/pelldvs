package chainflags

// common flags
var FromKeyNameFlag = &StringFlag{
	Name:  "from",
	Usage: "Name of the key to use for signing the transaction",
}

// EthRPCURLFlag is the flag for the Ethereum RPC URL of Pell EVM
var EthRPCURLFlag = &StringFlag{
	Name:  "rpc-url",
	Usage: "URL of the Ethereum RPC server",
}

var ChainIDFlag = &IntFlag{
	Name:  "chain-id",
	Usage: "Chain ID",
	Aliases: NewAliases(
		"chain_id",
	),
}

var GroupNumbers = &StringFlag{
	Name:  "groups",
	Usage: "group numbers",
}

var MetadataURI = &StringFlag{
	Name:  "metadata-uri",
	Usage: "metadata URI",
}

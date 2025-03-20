package chainflags

// flags for Pell EVM
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
	Usage: "Address of the DVS directory contract on Pell EVM",
}

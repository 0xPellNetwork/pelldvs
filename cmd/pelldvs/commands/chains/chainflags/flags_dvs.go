package chainflags

// flags for DVS

var DVSFrom = &StringFlag{
	Name:  "dvs-from",
	Usage: "dvs from key name",
}

var DVSRPCURL = &StringFlag{
	Name:  "dvs-rpc-url",
	Usage: "RPC URL of the DVS",
}

// TODO(menduo @2025-03-20, Thu, 14:10): set dvs-xxx as default name later
var DVSCentralSchedulerAddress = &StringFlag{
	Name:  "central-scheduler",
	Usage: "central scheduler address",
	Aliases: []string{
		"dvs-central-scheduler",
	},
}

var DVSOperatorStakeManagerAddress = &StringFlag{
	Name:  "dvs-operator-stake-manager",
	Usage: "Address of the DVS operator stake manager contract",
	Aliases: []string{
		"stake-registry",
	},
}

var DVSEjectionManagerAddress = &StringFlag{
	Name:  "dvs-ejection-manager",
	Usage: "ejection manager address",
	Aliases: []string{
		"ejection-manager",
	},
}

var DVSApproverKeyName = &StringFlag{
	Name:  "approver-key-name",
	Usage: "approver key name",
	Aliases: []string{
		"dvs-approver-key-name",
	},
}

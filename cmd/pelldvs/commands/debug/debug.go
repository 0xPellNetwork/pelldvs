package debug

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/log"
)

var (
	nodeRPCAddr string
	profAddr    string
	frequency   uint

	flagNodeRPCAddr = "rpc-laddr"
	flagProfAddr    = "pprof-laddr"
	flagFrequency   = "frequency"

	logger = log.NewLogger(os.Stdout)
)

// DebugCmd defines the root command containing subcommands that assist in
// debugging running PellDVS processes.
var DebugCmd = &cobra.Command{
	Use:   "debug",
	Short: "A utility to kill or watch a PellDVS process while aggregating debugging data",
}

func init() {
	DebugCmd.PersistentFlags().SortFlags = true
	DebugCmd.PersistentFlags().StringVar(
		&nodeRPCAddr,
		flagNodeRPCAddr,
		"tcp://localhost:26657",
		"the PellDVS node's RPC address (<host>:<port>)",
	)

	DebugCmd.AddCommand(killCmd)
	DebugCmd.AddCommand(dumpCmd)
}

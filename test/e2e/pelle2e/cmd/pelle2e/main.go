package main

import (
	"os"
	"path/filepath"

	"github.com/0xPellNetwork/pelldvs/libs/cli"
	"github.com/0xPellNetwork/pelldvs/test/e2e/pelle2e/cmd/pelle2e/commands"
)

func main() {
	rootCmd := commands.RootCmd
	rootCmd.AddCommand(
		commands.CheckBLSAggrSigCmd,
		commands.RequestDVSCmd,
	)
	cmd := cli.PrepareBaseCmd(rootCmd, "PELLDVS", os.ExpandEnv(filepath.Join("$HOME", ".pelldvs")))
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

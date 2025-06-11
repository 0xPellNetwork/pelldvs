package main

import (
	"os"
	"path/filepath"

	cmd "github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/debug"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/service"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/utils"
	"github.com/0xPellNetwork/pelldvs/libs/cli"
	nm "github.com/0xPellNetwork/pelldvs/node"
)

func main() {
	prompter := utils.NewPrompter()

	rootCmd := cmd.RootCmd
	rootCmd.AddCommand(
		cmd.GenValidatorCmd,
		cmd.InitFilesCmd,
		cmd.ShowValidatorCmd,
		cmd.ShowNodeIDCmd,
		cmd.GenNodeKeyCmd,
		cmd.VersionCmd,
		cmd.KeysCmd(prompter),
		client.ClientCmd,
		service.ServiceCmd,
		debug.DebugCmd,
		cmd.StartAggregatorCmd,
		cli.NewCompletionCmd(rootCmd, true),
	)

	// NOTE:
	// Users wishing to:
	//	* Use an external signer for their validators
	//	* Supply an in-proc abci app
	//	* Supply a genesis doc file from another source
	//	* Provide their own DB implementation
	// can copy this file and use something other than the
	// DefaultNewNode function
	nodeFunc := nm.DefaultNewNode
	//
	////Create & start node
	rootCmd.AddCommand(cmd.NewRunNodeCmd(nodeFunc))

	cmd := cli.PrepareBaseCmd(rootCmd, "PELLDVS", os.ExpandEnv(filepath.Join("$HOME", ".pelldvs")))
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

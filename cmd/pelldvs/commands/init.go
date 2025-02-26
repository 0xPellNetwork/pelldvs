package commands

import (
	"github.com/spf13/cobra"

	cfg "github.com/0xPellNetwork/pelldvs/config"
	cmtos "github.com/0xPellNetwork/pelldvs/libs/os"
	"github.com/0xPellNetwork/pelldvs/p2p"
)

// InitFilesCmd initializes a fresh PellDVS instance.
var InitFilesCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize PellDVS",
	RunE:  initFiles,
}

func initFiles(*cobra.Command, []string) error {
	return initFilesWithConfig(config)
}

func initFilesWithConfig(config *cfg.Config) error {
	// private validator
	privValKeyFile := config.PrivValidatorKeyFile()
	privValStateFile := config.PrivValidatorStateFile()
	// var pv *privval.FilePV
	if cmtos.FileExists(privValKeyFile) {
		// pv = privval.LoadFilePV(privValKeyFile, privValStateFile)
		logger.Info("Found private validator", "keyFile", privValKeyFile,
			"stateFile", privValStateFile)
	} else {
		// pv = privval.GenFilePV(privValKeyFile, privValStateFile)
		// pv.Save()
		logger.Info("Generated private validator", "keyFile", privValKeyFile,
			"stateFile", privValStateFile)
	}

	nodeKeyFile := config.NodeKeyFile()
	if cmtos.FileExists(nodeKeyFile) {
		logger.Info("Found node key", "path", nodeKeyFile)
	} else {
		if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
			return err
		}
		logger.Info("Generated node key", "path", nodeKeyFile)
	}

	// genesis file
	genFile := config.GenesisFile()
	if cmtos.FileExists(genFile) {
		logger.Info("Found genesis file", "path", genFile)
	} else {
		logger.Info("Generated genesis file", "path", genFile)
	}

	return nil
}

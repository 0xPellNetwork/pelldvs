package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	cmtjson "github.com/0xPellNetwork/pelldvs/libs/json"
	cmtos "github.com/0xPellNetwork/pelldvs/libs/os"
	"github.com/0xPellNetwork/pelldvs/privval"
)

// ShowValidatorCmd adds capabilities for showing the validator info.
var ShowValidatorCmd = &cobra.Command{
	Use:     "show-validator",
	Aliases: []string{"show_validator"},
	Short:   "Show this node's validator info",
	RunE:    showValidator,
}

func showValidator(*cobra.Command, []string) error {
	keyFilePath := config.PrivValidatorKeyFile()
	if !cmtos.FileExists(keyFilePath) {
		return fmt.Errorf("private validator file %s does not exist", keyFilePath)
	}

	pv := privval.LoadFilePV(config.Pell.OperatorBLSPrivateKeyStorePath)

	pubKey, err := pv.GetPubKey()
	if err != nil {
		return fmt.Errorf("can't get pubkey: %w", err)
	}

	bz, err := cmtjson.Marshal(pubKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private validator pubkey: %w", err)
	}

	fmt.Println(string(bz))
	return nil
}

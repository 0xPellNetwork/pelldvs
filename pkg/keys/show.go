package keys

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/utils"
)

func ShowCmd(p utils.Prompter) *cobra.Command {
	showCmd := &cobra.Command{
		Use:     "show",
		Short:   "show the keys of the given key name",
		Example: "show <key-name>",
		Aliases: []string{"s"},
		Long:    `This command will show the key details of the given key name`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyName := args[0]

			fmt.Println("Key Name: " + keyName)

			kps := GetKeysPath(pellcfg.CmtConfig, keyName)
			if !kps.IsAnyExists() {
				return fmt.Errorf("key does not exist: %s", keyName)
			}

			var avaliableTypes []string
			if kps.IsECDSAExist() {
				avaliableTypes = append(avaliableTypes, "ECDSA")
			}
			if kps.IsBLSExist() {
				avaliableTypes = append(avaliableTypes, "BLS")
			}

			fmt.Println("Available Key Types: " + fmt.Sprintf("%v", avaliableTypes))

			fmt.Println()
			if kps.IsECDSAExist() {
				fmt.Println("Key Name: " + keyName)
				fmt.Println("Key Type: ECDSA")
				keyFilePath := kps.ECDSA
				address, err := GetAddress(filepath.Clean(keyFilePath))
				if err != nil {
					return err
				}
				fmt.Println("Address: 0x" + address)
				fmt.Println("Key location: " + keyFilePath)

				// print file content
				keyContent, err := os.ReadFile(keyFilePath)
				if err != nil {
					return err
				}
				fmt.Println("Key content: ")
				fmt.Println(string(keyContent))

				fmt.Println("====================================================================================")
				fmt.Println()
			}

			if kps.IsBLSExist() {
				fmt.Println("Key Name: " + keyName)
				fmt.Println("Key Type: BLS")
				keyFilePath := kps.BLS
				pubKey, err := GetPubKey(filepath.Clean(keyFilePath))
				if err != nil {
					return err
				}
				fmt.Println("Public Key: " + pubKey)
				idStr, err := GetIDFromBLSPubKey(pubKey)
				if err != nil {
					return err
				}
				fmt.Println("Id: 0x" + idStr)
				fmt.Println("Key location: " + keyFilePath)

				// print file content
				keyContent, err := os.ReadFile(keyFilePath)
				if err != nil {
					return err
				}
				fmt.Println("Key content: ")
				fmt.Println(string(keyContent))

				fmt.Println("====================================================================================")
				fmt.Println()
			}
			return nil
		},
	}
	return showCmd
}

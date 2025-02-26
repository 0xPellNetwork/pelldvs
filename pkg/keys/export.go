package keys

import (
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/bls"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/pkg/utils"
)

func ExportCmd(p utils.Prompter) *cobra.Command {
	exportCmd := &cobra.Command{
		Use:     "export",
		Short:   "Used to export existing keys from local keystore",
		Example: "export --key-type <key-type> [flags] [keyname]",
		Aliases: []string{"e"},
		Long: `Used to export ecdsa and bls key from local keystore

keyname - This will be the name of the key to be imported. If the path of keys is
different from default path created by "create"/"import" command, then provide the
full path using --key-path flag.

If both keyname is provided and --key-path flag is provided, then keyname will be used.

use --key-type ecdsa/bls to export ecdsa/bls key.
- ecdsa - exported key should be plaintext hex encoded private key
- bls - exported key should be plaintext bls private key

It will prompt for password to encrypt the key.

This command will import keys from $HOME/.pelldvs/keys/ location

But if you want it to export from a different location, use --key-path flag`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyType := KeyTypeFlag.Value

			keyName := args[0]

			keyPath := KeyPathFlag.Value
			if len(keyPath) == 0 && len(keyName) == 0 {
				return errors.New("one of keyname or --key-path is required")
			}

			if len(keyPath) > 0 && len(keyName) > 0 {
				return errors.New("keyname and --key-path both are provided. Please provide only one")
			}

			filePath, err := getKeyPath(keyPath, keyName, keyType)
			if err != nil {
				return err
			}

			confirm, err := p.Confirm("This will show your private key. Are you sure you want to export?")
			if err != nil {
				return err
			}
			if !confirm {
				return nil
			}

			password, err := p.InputHiddenString("Enter password to decrypt the key", "", func(s string) error {
				return nil
			})
			if err != nil {
				return err
			}
			fmt.Println("exporting key from: ", filePath)

			privateKey, err := getPrivateKey(keyType, filePath, password)
			if err != nil {
				return err
			}
			fmt.Println("Private key: ", privateKey)
			return nil
		},
	}

	exportCmd.Flags().StringVar(&KeyTypeFlag.Value, KeyTypeFlag.Name, "", KeyTypeFlag.Usage)
	exportCmd.Flags().StringVar(&KeyPathFlag.Value, KeyPathFlag.Name, "", KeyPathFlag.Usage)

	return exportCmd
}

func getPrivateKey(keyType string, filePath string, password string) (string, error) {
	switch keyType {
	case KeyTypeECDSA:
		key, err := ecdsa.ReadKey(filePath, password)
		if err != nil {
			return "", err
		}
		return hex.EncodeToString(key.D.Bytes()), nil
	case KeyTypeBLS:
		key, err := bls.ReadPrivateKeyFromFile(filePath, password)
		if err != nil {
			return "", err
		}
		return key.PrivKey.String(), nil
	default:
		return "", ErrInvalidKeyType
	}
}

func getKeyPath(keyPath string, keyName string, keyType string) (string, error) {
	var filePath string
	if len(keyName) > 0 {
		switch keyType {
		case KeyTypeECDSA:
			filePath = GetKeysPath(pellcfg.CmtConfig, keyName).ECDSA
		case KeyTypeBLS:
			filePath = GetKeysPath(pellcfg.CmtConfig, keyName).BLS
		default:
			return "", ErrInvalidKeyType
		}

	} else {
		filePath = filepath.Clean(keyPath)
	}

	return filePath, nil
}

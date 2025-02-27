package keys

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/bls"
	"github.com/0xPellNetwork/pelldvs/pkg/utils"
)

func ImportCmd(p utils.Prompter) *cobra.Command {
	importCmd := &cobra.Command{
		Use:     "import",
		Short:   "Used to import existing keys in local keystore",
		Example: "import --key-type <key-type> [flags] <keyname> <private-key>",
		Aliases: []string{"i"},
		Long: `
Used to import ecdsa and bls key in local keystore

keyname (required) - This will be the name of the imported key file. It will be saved as <keyname>.ecdsa.key.json or <keyname>.bls.key.json

use --key-type ecdsa/bls to import ecdsa/bls key.
- ecdsa - <private-key> should be plaintext hex encoded private key
- bls - <private-key> should be plaintext bls private key

It will prompt for password to encrypt the key, which is optional but highly recommended.
If you want to import a key with weak/no password, use --insecure flag. Do NOT use those keys in production

This command also support piping the password from stdin.
For example: echo "password" | pelldvs keys import --key-type ecdsa keyname privateKey

This command will import keys in $HOME/.pelldvs/keys/ location
		`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			keyName := args[0]
			if err := validateKeyName(keyName); err != nil {
				return err
			}

			privateKey := args[1]
			if err := validatePrivateKey(privateKey); err != nil {
				return err
			}

			// Check if input is available in the pipe and read the password from it
			stdInPassword, readFromPipe := utils.GetStdInPassword()

			keyType := KeyTypeFlag.Value
			insecure := InsecureFlag.Value

			switch keyType {
			case KeyTypeECDSA:
				privateKey = strings.TrimPrefix(privateKey, "0x")
				privateKeyPair, err := crypto.HexToECDSA(privateKey)
				if err != nil {
					return err
				}
				return saveEcdsaKey(keyName, p, privateKeyPair, insecure, stdInPassword, readFromPipe)
			case KeyTypeBLS:
				privateKeyBigInt := new(big.Int)
				_, ok := privateKeyBigInt.SetString(privateKey, 10)
				var blsKeyPair *bls.KeyPair
				var err error
				if ok {
					fmt.Println("Importing from large integer")
					blsKeyPair, err = bls.NewKeyPairFromString(privateKey)
					if err != nil {
						return err
					}
				} else {
					// Try to parse as hex
					fmt.Println("Importing from hex")
					z := new(big.Int)
					privateKey = strings.TrimPrefix(privateKey, "0x")
					_, ok := z.SetString(privateKey, 16)
					if !ok {
						return ErrInvalidHexPrivateKey
					}
					blsKeyPair, err = bls.NewKeyPairFromString(z.String())
					if err != nil {
						return err
					}
				}
				return saveBlsKey(keyName, p, blsKeyPair, insecure, stdInPassword, readFromPipe)
			default:
				return ErrInvalidKeyType
			}
		},
	}

	importCmd.Flags().StringVarP(&KeyTypeFlag.Value, KeyTypeFlag.Name, KeyTypeFlag.Aliases, "", "Type of key to import (ecdsa/bls)")
	importCmd.Flags().BoolVarP(&InsecureFlag.Value, InsecureFlag.Name, InsecureFlag.Aliases, false, "Import key without password")

	return importCmd
}

func validatePrivateKey(pk string) error {
	if len(pk) == 0 {
		return ErrEmptyPrivateKey
	}

	if match, _ := regexp.MatchString("\\s", pk); match {
		return ErrPrivateKeyContainsWhitespaces
	}

	return nil
}

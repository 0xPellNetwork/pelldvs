package keys

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	passwordvalidator "github.com/wagslane/go-password-validator"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/bls"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	sdkEcdsa "github.com/0xPellNetwork/pelldvs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/pkg/utils"
)

const (
	KeyTypeECDSA = "ecdsa"
	KeyTypeBLS   = "bls"

	// MinEntropyBits For password validation
	MinEntropyBits = 70
)

func CreateCmd(p utils.Prompter) *cobra.Command {
	createCmd := &cobra.Command{
		Use:     "create",
		Short:   "Used to create encrypted keys in local keystore",
		Example: "create --key-type <key-type> [flags] <keyname>",
		Aliases: []string{"c"},
		Long: `
Used to create ecdsa and bls key in local keystore

keyname (required) - This will be the name of the created key file. It will be saved as <keyname>.ecdsa.key.json or <keyname>.bls.key.json

use --key-type ecdsa/bls to create ecdsa/bls key.
It will prompt for password to encrypt the key, which is optional but highly recommended.
If you want to create a key with weak/no password, use --insecure flag. Do NOT use those keys in production

This command also support piping the password from stdin.
For example: echo "password" | pelldvs keys create --key-type ecdsa keyname

This command will create keys in $HOME/.pelldvs/keys/ location
		`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyName := args[0]
			if err := validateKeyName(keyName); err != nil {
				return err
			}

			// Check if input is available in the pipe and read the password from it
			stdInPassword, readFromPipe := utils.GetStdInPassword()

			keyType := KeyTypeFlag.Value
			insecure := InsecureFlag.Value

			switch keyType {
			case KeyTypeECDSA:
				privateKey, err := crypto.GenerateKey()
				if err != nil {
					return err
				}
				return saveEcdsaKey(keyName, p, privateKey, insecure, stdInPassword, readFromPipe)
			case KeyTypeBLS:
				blsKeyPair, err := bls.GenRandomBlsKeys()
				if err != nil {
					return err
				}
				return saveBlsKey(keyName, p, blsKeyPair, insecure, stdInPassword, readFromPipe)
			default:
				return ErrInvalidKeyType
			}
		},
	}

	createCmd.Flags().StringVarP(&KeyTypeFlag.Value, KeyTypeFlag.Name, KeyTypeFlag.Aliases, "", "Type of key to create (ecdsa/bls)")
	createCmd.Flags().BoolVarP(&InsecureFlag.Value, InsecureFlag.Name, InsecureFlag.Aliases, false, "Create key without password")

	return createCmd
}

func validateKeyName(keyName string) error {
	if len(keyName) == 0 {
		return ErrEmptyKeyName
	}

	if match, _ := regexp.MatchString("\\s", keyName); match {
		return ErrKeyContainsWhitespaces
	}

	return nil
}

func saveBlsKey(
	keyName string,
	p utils.Prompter,
	keyPair *bls.KeyPair,
	insecure bool,
	stdInPassword string,
	readFromPipe bool,
) error {
	var err error
	fileLoc := GetKeysPath(pellcfg.CmtConfig, keyName).BLS

	if checkIfKeyExists(fileLoc) {
		return errors.New("key name already exists. Please choose a different name")
	}

	var password string
	if !readFromPipe {
		password, err = getPasswordFromPrompt(p, insecure, "Enter password to encrypt the bls private key:")
		if err != nil {
			return err
		}
	} else {
		password = stdInPassword
		if !insecure {
			err = validatePassword(password)
			if err != nil {
				return err
			}
		}
	}

	err = keyPair.SaveToFile(fileLoc, password)
	if err != nil {
		return err
	}

	privateKeyHex := keyPair.PrivKey.String()
	publicKeyHex := keyPair.PubKey.String()

	fmt.Printf("\nKey location: %s\nPublic Key: %s\n\n", fileLoc, publicKeyHex)
	return displayWithLess(privateKeyHex, KeyTypeBLS)
}

func saveEcdsaKey(
	keyName string,
	p utils.Prompter,
	privateKey *ecdsa.PrivateKey,
	insecure bool,
	stdInPassword string,
	readFromPipe bool,
) error {
	var err error
	fileLoc := GetKeysPath(pellcfg.CmtConfig, keyName).ECDSA

	if checkIfKeyExists(fileLoc) {
		return errors.New("key name already exists. Please choose a different name")
	}

	var password string
	if !readFromPipe {
		password, err = getPasswordFromPrompt(p, insecure, "Enter password to encrypt the ecdsa private key:")
		if err != nil {
			return err
		}
	} else {
		password = stdInPassword
		if !insecure {
			err = validatePassword(password)
			if err != nil {
				return err
			}
		}
	}

	err = sdkEcdsa.WriteKey(fileLoc, privateKey, password)
	if err != nil {
		return err
	}

	privateKeyHex := hex.EncodeToString(privateKey.D.Bytes())

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("error casting public key to ECDSA public key")
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	publicKeyHex := hexutil.Encode(publicKeyBytes)[4:]
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	fmt.Printf("\nKey location: %s\nPublic Key hex: %s\nEthereum Address: %s\n\n", fileLoc, publicKeyHex, address)
	return displayWithLess(privateKeyHex, KeyTypeECDSA)
}

func padLeft(str string, length int) string {
	for len(str) < length {
		str = "0" + str
	}
	return str
}

func displayWithLess(privateKeyHex string, keyType string) error {
	var message, border, keyLine string
	tabSpace := "    "

	// Pad with 0 to match size of 64 bytes
	if keyType == KeyTypeECDSA {
		privateKeyHex = padLeft(privateKeyHex, 64)
	}
	keyContent := tabSpace + privateKeyHex + tabSpace
	borderLength := len(keyContent) + 4
	border = strings.Repeat("/", borderLength)
	paddingLine := "//" + strings.Repeat(" ", borderLength-4) + "//"

	keyLine = fmt.Sprintf("//%s//", keyContent)

	if keyType == KeyTypeECDSA {
		message = fmt.Sprintf(`
ECDSA Private Key (Hex):

%s
%s
%s
%s
%s

🔐 Please backup the above private key hex in a safe place 🔒

`, border, paddingLine, keyLine, paddingLine, border)
	} else if keyType == KeyTypeBLS {
		message = fmt.Sprintf(`
BLS Private Key (Hex):

%s
%s
%s
%s
%s

🔐 Please backup the above private key hex in a safe place 🔒

`, border, paddingLine, keyLine, paddingLine, border)
	}

	cmd := exec.Command("less", "-R")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("error creating stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting less command: %w", err)
	}

	if _, err := stdin.Write([]byte(message)); err != nil {
		return fmt.Errorf("error writing message to less command: %w", err)
	}

	if err := stdin.Close(); err != nil {
		return fmt.Errorf("error closing stdin pipe: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error waiting for less command: %w", err)
	}

	return nil
}

func getPasswordFromPrompt(p utils.Prompter, insecure bool, prompt string) (string, error) {
	password, err := p.InputHiddenString(prompt, "",
		func(s string) error {
			if insecure {
				return nil
			}
			return validatePassword(s)
		},
	)
	if err != nil {
		return "", err
	}
	_, err = p.InputHiddenString("Please confirm your password:", "",
		func(s string) error {
			if s != password {
				return errors.New("passwords are not matched")
			}
			return nil
		},
	)
	if err != nil {
		return "", err
	}
	return password, nil
}

func checkIfKeyExists(fileLoc string) bool {
	_, err := os.Stat(fileLoc)
	return !os.IsNotExist(err)
}

func validatePassword(password string) error {
	err := passwordvalidator.Validate(password, MinEntropyBits)
	if err != nil {
		fmt.Println(
			"if you want to create keys for testing with weak/no password, use --insecure flag. Do NOT use those keys in production",
		)
	}
	return err
}

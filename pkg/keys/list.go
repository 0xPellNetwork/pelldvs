package keys

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/bls"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/utils"
	"github.com/0xPellNetwork/pelldvs/types"
)

func ListCmd(p utils.Prompter) *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "List all the keys created by this create command",
		Example: "list",
		Aliases: []string{"l"},
		Long: `
This command will list both ecdsa and bls key created using create command

It will only list keys created in the default folder (./keys/)
		`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyStorePath := GetKeysStoredDir(pellcfg.CmtConfig)
			files, err := os.ReadDir(keyStorePath)
			if err != nil {
				return err
			}

			for _, file := range files {
				// if file is a directory, skip it
				if file.IsDir() {
					fmt.Println("Skipping directory: " + file.Name())
					continue
				}

				keySplits := strings.Split(file.Name(), ".")
				if len(keySplits) <= 2 {
					fmt.Println("Invalid filename, skipping : " + file.Name())
					continue
				}

				fileName := keySplits[0]
				keyType := keySplits[1]
				fmt.Println("Key Name: " + fileName)
				switch keyType {
				case KeyTypeECDSA:
					fmt.Println("Key Type: ECDSA")
					keyFilePath := filepath.Join(keyStorePath, file.Name())
					address, err := GetAddress(filepath.Clean(keyFilePath))
					if err != nil {
						return err
					}
					fmt.Println("Address: 0x" + address)
					fmt.Println("Key location: " + keyFilePath)
					fmt.Println("====================================================================================")
					fmt.Println()
				case KeyTypeBLS:
					fmt.Println("Key Type: BLS")
					keyFilePath := filepath.Join(keyStorePath, file.Name())
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
					fmt.Println("====================================================================================")
					fmt.Println()
				}

			}
			return nil
		},
	}
	return listCmd
}

func GetPubKey(keyStoreFile string) (string, error) {
	keyJSON, err := os.ReadFile(keyStoreFile)
	if err != nil {
		return "", err
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(keyJSON, &m); err != nil {
		return "", err
	}

	if pubKey, ok := m["pubKey"].(string); !ok {
		return "", fmt.Errorf("pubKey not found in key file")
	} else {
		return pubKey, nil
	}
}

func GetIDFromBLSPubKey(pubKey string) (string, error) {
	// The pubkey 's string is generated from this code:
	// ```go
	// func (p *G1Affine) String() string {
	// 	if p.IsInfinity() {
	// 		return "O"
	//	}
	//	return "E([" + p.X.String() + "," + p.Y.String() + "])"
	// }
	// ```
	//
	// This code just parser this string:
	// E([498211989701534593628498974128726712526336918939770789545660245177948853517,19434346619705907282579203143605058653932187676054178921788041096426532277474])

	if pubKey == "O" {
		return "", fmt.Errorf("pubKey is Infinity")
	}

	if pubKey[:3] != "E([" && pubKey[len(pubKey)-2:] != "])" {
		return "", fmt.Errorf("pubKey format failed by not E([x,y])")
	}

	pubKeyStr := pubKey[3 : len(pubKey)-2]
	strs := strings.Split(pubKeyStr, ",")
	if len(strs) != 2 {
		return "", fmt.Errorf("pubkey format failed by not x,y")
	}

	xe, err := new(fp.Element).SetString(strs[0])
	if err != nil {
		return "", err
	}

	ye, err := new(fp.Element).SetString(strs[1])
	if err != nil {
		return "", err
	}

	point := &bls.G1Point{
		G1Affine: &bn254.G1Affine{
			X: *xe,
			Y: *ye,
		},
	}

	id := types.GenOperatorIDOffChain(point)

	return id.LogValue().String(), nil
}

func GetAddress(keyStoreFile string) (string, error) {
	keyJSON, err := os.ReadFile(keyStoreFile)
	if err != nil {
		return "", err
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(keyJSON, &m); err != nil {
		return "", err
	}

	if address, ok := m["address"].(string); !ok {
		return "", fmt.Errorf("address not found in key file")
	} else {
		return address, nil
	}
}

// GetECDSAPrivateKey - Keeping it right now as we might need this function to export
// the keys
func GetECDSAPrivateKey(keyStoreFile string, password string) (*ecdsa.PrivateKey, error) {
	keyStoreContents, err := os.ReadFile(keyStoreFile)
	if err != nil {
		return nil, err
	}

	sk, err := keystore.DecryptKey(keyStoreContents, password)
	if err != nil {
		return nil, err
	}

	return sk.PrivateKey, nil
}

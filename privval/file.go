package privval

import (
	"fmt"

	"github.com/0xPellNetwork/pelldvs/crypto"
	"github.com/0xPellNetwork/pelldvs/crypto/bls"
	cmtos "github.com/0xPellNetwork/pelldvs/libs/os"
)

type Address = crypto.Address

const (
	emptyPassword = ""
)

type FilePVKey struct {
	KeyPair bls.KeyPair `json:"key_pair"`

	filePath string
}

func (v FilePVKey) GetKeyPair() bls.KeyPair {
	return v.KeyPair
}

func (v FilePVKey) Save() {
	if v.filePath == "" {
		panic("cannot save FilePVKey: filePath not set")
	}
	err := v.KeyPair.SaveToFile(v.filePath, emptyPassword)
	if err != nil {
		panic(err)
	}
}

//-------------------------------------------------------------------------------

// FilePV implements PrivValidator using data persisted to disk
// to prevent double signing.
// NOTE: the directories containing pv.Key.filePath and pv.LastSignState.filePath must already exist.
// It includes the LastSignature and LastSignBytes so we don't lose the signature
// if the process crashes after signing but before the resulting consensus message is processed.
type FilePV struct {
	Key FilePVKey
}

// NewFilePV generates a new validator from the given key and paths.
func NewFilePV(blsKeyPair bls.KeyPair, blsKeyFilePath string) *FilePV {
	return &FilePV{
		Key: FilePVKey{
			KeyPair:  blsKeyPair,
			filePath: blsKeyFilePath,
		},
	}
}

// GenFilePV generates a new validator with randomly generated private key
// and sets the filePaths, but does not call Save().
func GenFilePV(blsKeyFilePath string) (*FilePV, error) {
	blsKeys, err := bls.GenRandomBlsKeys()
	if err != nil {
		return nil, err
	}
	return NewFilePV(*blsKeys, blsKeyFilePath), nil
}

// LoadFilePV loads a FilePV from the filePaths.  The FilePV handles double
// signing prevention by persisting data to the stateFilePath.  If either file path
// does not exist, the program will exit.
func LoadFilePV(keyFilePath string) *FilePV {
	return loadFilePV(keyFilePath)
}

// If loadState is true, we load from the stateFilePath. Otherwise, we use an empty LastSignState.
func loadFilePV(keyFilePath string) *FilePV {
	// Load private key from file
	privateKey, err := bls.ReadPrivateKeyFromFile(keyFilePath, emptyPassword)
	if err != nil {
		cmtos.Exit(fmt.Sprintf("Error reading BLS private key from %v: %v\n", keyFilePath, err))
	}

	// Create and return FilePVKey instance
	blsKey := FilePVKey{
		KeyPair:  *privateKey,
		filePath: keyFilePath,
	}

	return &FilePV{
		Key: blsKey,
	}
}

// LoadOrGenFilePV loads a FilePV from the given filePaths
// or else generates a new one and saves it to the filePaths.
func LoadOrGenFilePV(keyFilePath string) (*FilePV, error) {
	var pv *FilePV
	var err error
	if cmtos.FileExists(keyFilePath) {
		pv = LoadFilePV(keyFilePath)
	} else {
		pv, err = GenFilePV(keyFilePath)
		if err != nil {
			return nil, err
		}
		pv.Save()
	}
	return pv, nil
}

func (v *FilePV) SignBytes(bytes []byte) (*bls.Signature, error) {
	var msg [32]byte
	copy(msg[:], bytes)
	pair := v.Key.GetKeyPair()
	sig := pair.SignMessage(msg)
	return sig, nil
}

// GetPubKey returns the public key of the validator.
// Implements PrivValidator.
func (pv *FilePV) GetPubKey() (*bls.G1Point, error) {
	return pv.Key.KeyPair.PubKey, nil
}

// Save persists the FilePV to disk.
func (pv *FilePV) Save() {
	pv.Key.Save()
}

// String returns a string representation of the FilePV.
func (pv *FilePV) String() string {
	return fmt.Sprintf("FilePV{%v}", pv.Key.KeyPair.PubKey)
}

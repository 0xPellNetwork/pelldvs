package keys

import (
	"fmt"

	cmtcfg "github.com/0xPellNetwork/pelldvs/config"
)

func GetKeysStoredDir(cfg *cmtcfg.Config) string {
	return cfg.RootDir + "/keys"
}

type KeyPath struct {
	ECDSA string
	BLS   string
}

func (kp *KeyPath) IsECDSAExist() bool {
	if kp.ECDSA == "" {
		return false
	}
	return checkIfKeyExists(kp.ECDSA)
}

func (kp *KeyPath) IsBLSExist() bool {
	if kp.BLS == "" {
		return false
	}
	return checkIfKeyExists(kp.BLS)
}

func (kp *KeyPath) IsAllExists() bool {
	return kp.IsECDSAExist() && kp.IsBLSExist()
}

func (kp *KeyPath) IsAnyExists() bool {
	return kp.IsECDSAExist() || kp.IsBLSExist()
}

func GetKeysPath(cfg *cmtcfg.Config, name string) KeyPath {
	keysDir := GetKeysStoredDir(cfg)
	return KeyPath{
		ECDSA: fmt.Sprintf("%s/%s.ecdsa.key.json", keysDir, name),
		BLS:   fmt.Sprintf("%s/%s.bls.key.json", keysDir, name),
	}
}

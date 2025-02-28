package utils

import (
	"strings"

	gethcommon "github.com/ethereum/go-ethereum/common"

	chaincfg "github.com/0xPellNetwork/pelldvs-interactor/config"
)

func NewChainConfigChecker(cfg *chaincfg.Config) *ChainConifigChecker {
	var res = &ChainConifigChecker{cfg: cfg}
	return res
}

type ChainConifigChecker struct {
	cfg *chaincfg.Config
}

func (cc *ChainConifigChecker) HasRPCURL() bool {
	return cc.cfg != nil && cc.cfg.RPCURL != ""
}

func (cc *ChainConifigChecker) IsValidPellRegistryRouterFactory() bool {
	return cc.cfg != nil &&
		cc.isValidNoZeroAddress(cc.cfg.ContractConfig.PellRegistryRouterFactory)
}

func (cc *ChainConifigChecker) IsValidPellRegistryRouter() bool {
	return cc.cfg != nil &&
		cc.isValidNoZeroAddress(cc.cfg.ContractConfig.PellRegistryRouter)
}

func (cc *ChainConifigChecker) IsValidPellDelegationManager() bool {
	return cc.cfg != nil &&
		cc.isValidNoZeroAddress(cc.cfg.ContractConfig.PellDelegationManager)
}

func (cc *ChainConifigChecker) IsValidPellDVSDirectory() bool {
	return cc.cfg != nil &&
		cc.isValidNoZeroAddress(cc.cfg.ContractConfig.PellDVSDirectory)
}

func (cc *ChainConifigChecker) HasECDSAPrivateKeyFilePath() bool {
	return cc.cfg != nil &&
		cc.cfg.ECDSAPrivateKeyFilePath != ""
}

func (cc *ChainConifigChecker) HasDVSRPCURL(chainID uint64) bool {
	return cc.cfg != nil &&
		cc.cfg.ContractConfig.DVSConfigs != nil &&
		cc.cfg.ContractConfig.DVSConfigs[chainID] != nil &&
		cc.cfg.ContractConfig.DVSConfigs[chainID].RPCURL != ""
}

func (cc *ChainConifigChecker) IsValidDVSCentralScheduler(chainID uint64) bool {
	return cc.cfg != nil &&
		cc.cfg.ContractConfig.DVSConfigs != nil &&
		cc.cfg.ContractConfig.DVSConfigs[chainID] != nil &&
		cc.isValidNoZeroAddress(cc.cfg.ContractConfig.DVSConfigs[chainID].CentralScheduler)

}

func (cc *ChainConifigChecker) IsValidDVSOperatorInfoProvider(chainID uint64) bool {
	return cc.cfg != nil &&
		cc.cfg.ContractConfig.DVSConfigs != nil &&
		cc.cfg.ContractConfig.DVSConfigs[chainID] != nil &&
		cc.isValidNoZeroAddress(cc.cfg.ContractConfig.DVSConfigs[chainID].OperatorInfoProvider)
}

func (cc *ChainConifigChecker) IsValidDVSEjectionManager(chainID uint64) bool {
	return cc.cfg != nil &&
		cc.cfg.ContractConfig.DVSConfigs != nil &&
		cc.cfg.ContractConfig.DVSConfigs[chainID] != nil &&
		cc.isValidNoZeroAddress(cc.cfg.ContractConfig.DVSConfigs[chainID].EjectionManager)
}

func (cc *ChainConifigChecker) IsValidDVSOperatorStakeManager(chainID uint64) bool {
	return cc.cfg != nil &&
		cc.cfg.ContractConfig.DVSConfigs != nil &&
		cc.cfg.ContractConfig.DVSConfigs[chainID] != nil &&
		cc.isValidNoZeroAddress(cc.cfg.ContractConfig.DVSConfigs[chainID].OperatorStakeManager)
}

func (cc *ChainConifigChecker) isValidNoZeroAddress(s string) bool {
	return strings.TrimSpace(s) != "" && gethcommon.IsHexAddress(s) && gethcommon.HexToAddress(s) != gethcommon.Address{}
}

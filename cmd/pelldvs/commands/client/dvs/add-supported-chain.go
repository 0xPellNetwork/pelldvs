package dvs

import (
	"context"
	osecdsa "crypto/ecdsa"
	"fmt"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	chaincfg "github.com/0xPellNetwork/pelldvs-interactor/config"
	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

var (
	addSupportedChainCmdFlagCentralScheduler = &chainflags.StringFlag{
		Name:  "central-scheduler",
		Usage: "central scheduler address",
	}
	addSupportedChainCmdFlagEjectionManager = &chainflags.StringFlag{
		Name:  "ejection-manager",
		Usage: "ejection manager address",
	}
	addSupportedChainCmdFlagStakeRegistry = &chainflags.StringFlag{
		Name:  "stake-registry",
		Usage: "stake registry address",
	}
	addSupportedChainCmdFlagDVSRPCURL = &chainflags.StringFlag{
		Name:  "dvs-rpc-url",
		Usage: "dvs rpc url",
	}
	addSupportedChainCmdFlagApproverKeyName = &chainflags.StringFlag{
		Name:  "approver-key-name",
		Usage: "approver key name",
	}
)

func init() {
	chainflags.ChainIDFlag.AddToCmdFlag(addSupportedChainCmd)
	addSupportedChainCmdFlagCentralScheduler.AddToCmdFlag(addSupportedChainCmd)
	addSupportedChainCmdFlagEjectionManager.AddToCmdFlag(addSupportedChainCmd)
	addSupportedChainCmdFlagStakeRegistry.AddToCmdFlag(addSupportedChainCmd)
	addSupportedChainCmdFlagDVSRPCURL.AddToCmdFlag(addSupportedChainCmd)
	addSupportedChainCmdFlagApproverKeyName.AddToCmdFlag(addSupportedChainCmd)

	err := chainflags.MarkFlagsAreRequired(addSupportedChainCmd,
		chainflags.ChainIDFlag,
		addSupportedChainCmdFlagCentralScheduler,
		addSupportedChainCmdFlagEjectionManager,
		addSupportedChainCmdFlagStakeRegistry,
		addSupportedChainCmdFlagDVSRPCURL,
		addSupportedChainCmdFlagApproverKeyName,
	)
	if err != nil {
		panic(err)
	}
}

// handleAddSupportedChain command
var addSupportedChainCmd = &cobra.Command{
	Use:   "add-supported-chain",
	Short: "add supported chain",
	Long:  `add supported chain`,
	Example: `
pelldvs client dvs add-supported-chain \
	--from <key-name> \
	--rpc-url <rpc-url> \
	--registry-router <registry-router> \
	--chain-id <chain-id> \
	--central-scheduler <central-scheduler> \
	--ejection-manager <ejection-manager> \
	--stake-registry <stake-registry> \
	--dvs-rpc-url <dvs-rpc-url> \
	--approver-key-name <approver-key-name>

pelldvs client dvs add-supported-chain \
	--from pell-localnet-deployer \
	--rpc-url http://localhost:8545 \
	--registry-router "0xE7402c51ae0bd667ad57a99991af6C2b686cd4f1" \
	--chain-id 1337 \
	--central-scheduler "0x04C89607413713Ec9775E14b954286519d836FEf" \
	--ejection-manager "0x0355B7B8cb128fA5692729Ab3AAa199C1753f726" \
	--stake-registry "0x2E2Ed0Cfd3AD2f1d34481277b3204d807Ca2F8c2" \
	--dvs-rpc-url http://localhost:8545 \
	--approver-key-name pell-localnet-deployer

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleAddSupportedChain(cmd, chainflags.FromKeyNameFlag.Value, addSupportedChainCmdFlagApproverKeyName.Value)
	},
}

func handleAddSupportedChain(cmd *cobra.Command, keyName string, approverKeyName string) error {
	kpath := keys.GetKeysPath(pellcfg.CmtConfig, keyName)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	appRoverKeyPath := keys.GetKeysPath(pellcfg.CmtConfig, approverKeyName)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", appRoverKeyPath.ECDSA)
	}

	approverPkPair, err := ecdsa.ReadKey(appRoverKeyPath.ECDSA, "")
	if err != nil {
		return fmt.Errorf("failed to read approverPkPair ecdsa key: %v", err)
	}

	receipt, err := execAddSupportedChain(cmd, uint64(chainflags.ChainIDFlag.Value), kpath.ECDSA, approverPkPair)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execAddSupportedChain(cmd *cobra.Command, chainID uint64, privKeyPath string, appriverPk *osecdsa.PrivateKey) (*gethtypes.Receipt, error) {
	cmdName := "handleAddSupportedChain"

	logger.Info(fmt.Sprintf("%s start", cmdName),
		"privKeyPath", privKeyPath,
		"chainId", chainID,
	)

	ctx := context.Background()
	senderAddress, err := ecdsa.GetAddressFromKeyStoreFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get senderAddress from key store file: %v", err)
	}
	logger.Info(cmdName,
		"sender", senderAddress,
	)

	cfg, err := utils.LoadChainConfig(cmd, pellcfg.CmtConfig.Pell.InteractorConfigPath, logger)
	if err != nil {
		return nil, err
	}

	if cfg.ContractConfig.DVSConfigs == nil {
		cfg.ContractConfig.DVSConfigs = make(map[uint64]*chaincfg.DVSConfig)
	}
	if cfg.ContractConfig.DVSConfigs[chainID] == nil {
		cfg.ContractConfig.DVSConfigs[chainID] = &chaincfg.DVSConfig{}
	}

	if addSupportedChainCmdFlagCentralScheduler.Value != "" {
		cfg.ContractConfig.DVSConfigs[chainID].CentralScheduler = addSupportedChainCmdFlagCentralScheduler.Value
	}
	if addSupportedChainCmdFlagEjectionManager.Value != "" {
		cfg.ContractConfig.DVSConfigs[chainID].EjectionManager = addSupportedChainCmdFlagEjectionManager.Value
	}
	if addSupportedChainCmdFlagStakeRegistry.Value != "" {
		cfg.ContractConfig.DVSConfigs[chainID].OperatorStakeManager = addSupportedChainCmdFlagStakeRegistry.Value
	}
	if addSupportedChainCmdFlagDVSRPCURL.Value != "" {
		cfg.ContractConfig.DVSConfigs[chainID].RPCURL = addSupportedChainCmdFlagDVSRPCURL.Value
	}

	logger.Debug("chaincfg", "cfg", fmt.Sprintf("%+v", cfg))

	chainDVS, err := utils.NewDVSFromCfg(cfg, logger)
	if err != nil {
		return nil, err
	}

	receipt, err := chainDVS.AddSuportedChain(ctx, appriverPk, chainID)
	if err != nil {
		return nil, err
	}

	logger.Info(
		fmt.Sprintf("%s done", cmdName),
		"k", "v",
		"senderAddress", senderAddress,
		"receipt", receipt,
	)

	return receipt, err
}

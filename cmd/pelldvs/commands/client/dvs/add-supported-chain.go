package dvs

import (
	"context"
	osecdsa "crypto/ecdsa"
	"fmt"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-interactor/chainlibs/eth"
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
	addSupportedChainCmdFlagCentralScheduler.AddToCmdFlag(addSupportedChainCmd)
	addSupportedChainCmdFlagEjectionManager.AddToCmdFlag(addSupportedChainCmd)
	addSupportedChainCmdFlagStakeRegistry.AddToCmdFlag(addSupportedChainCmd)
	addSupportedChainCmdFlagDVSRPCURL.AddToCmdFlag(addSupportedChainCmd)
	addSupportedChainCmdFlagApproverKeyName.AddToCmdFlag(addSupportedChainCmd)

	err := chainflags.MarkFlagsAreRequired(addSupportedChainCmd,
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
	Long: `
pelldvs client dvs add-supported-chain \
	--from <key-name> \
	--rpc-url <rpc-url> \
	--registry-router <registry-router> \
	--central-scheduler <central-scheduler> \
	--ejection-manager <ejection-manager> \
	--stake-registry <stake-registry> \
	--dvs-rpc-url <dvs-rpc-url> \
	--approver-key-name <approver-key-name>
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleAddSupportedChain(cmd)
	},
}

func handleAddSupportedChain(cmd *cobra.Command) error {
	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	appRoverKeyPath := keys.GetKeysPath(pellcfg.CmtConfig, addSupportedChainCmdFlagApproverKeyName.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", appRoverKeyPath.ECDSA)
	}

	approverPkPair, err := ecdsa.ReadKey(appRoverKeyPath.ECDSA, "")
	if err != nil {
		return fmt.Errorf("failed to read approverPkPair ecdsa key: %v", err)
	}

	dvsETHClient, err := eth.NewClient(addSupportedChainCmdFlagDVSRPCURL.Value)
	if err != nil {
		return fmt.Errorf("failed to create eth client for RPCURL: %s:, err %v",
			registerChainToPellCmdFlagDVSRPCURL.Value, err,
		)
	}
	dvsChainID, err := dvsETHClient.ChainID(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get chain id from rpc client: url:%s , error: %v",
			registerChainToPellCmdFlagDVSRPCURL.Value, err)
	}

	receipt, err := execAddSupportedChain(cmd, dvsChainID.Uint64(), kpath.ECDSA, approverPkPair)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execAddSupportedChain(cmd *cobra.Command, dvsChainID uint64, privKeyPath string, approverPK *osecdsa.PrivateKey) (*gethtypes.Receipt, error) {
	cmdName := utils.GetPrettyCommandName(cmd)

	logger.Info(fmt.Sprintf("%s start", cmdName),
		"privKeyPath", privKeyPath,
		"chainId", dvsChainID,
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
		logger.Error("failed to load chain config", "err", err, "file", pellcfg.CmtConfig.Pell.InteractorConfigPath)
		return nil, err
	}

	if cfg.ContractConfig.DVSConfigs == nil {
		cfg.ContractConfig.DVSConfigs = make(map[uint64]*chaincfg.DVSConfig)
	}
	if cfg.ContractConfig.DVSConfigs[dvsChainID] == nil {
		cfg.ContractConfig.DVSConfigs[dvsChainID] = &chaincfg.DVSConfig{}
	}

	if addSupportedChainCmdFlagCentralScheduler.Value != "" {
		cfg.ContractConfig.DVSConfigs[dvsChainID].CentralScheduler = addSupportedChainCmdFlagCentralScheduler.Value
	}
	if addSupportedChainCmdFlagEjectionManager.Value != "" {
		cfg.ContractConfig.DVSConfigs[dvsChainID].EjectionManager = addSupportedChainCmdFlagEjectionManager.Value
	}
	if addSupportedChainCmdFlagStakeRegistry.Value != "" {
		cfg.ContractConfig.DVSConfigs[dvsChainID].OperatorStakeManager = addSupportedChainCmdFlagStakeRegistry.Value
	}
	if addSupportedChainCmdFlagDVSRPCURL.Value != "" {
		cfg.ContractConfig.DVSConfigs[dvsChainID].RPCURL = addSupportedChainCmdFlagDVSRPCURL.Value
	}

	logger.Info("chain config details", "chaincfg", fmt.Sprintf("%+v", cfg))

	var chainConfigChecker = utils.NewChainConfigChecker(cfg)
	if !chainConfigChecker.HasRPCURL() {
		return nil, fmt.Errorf("rpc url is required")
	}
	if !chainConfigChecker.IsValidPellRegistryRouter() {
		return nil, fmt.Errorf("pell registry router is required")
	}

	if !chainConfigChecker.IsValidDVSCentralScheduler(dvsChainID) {
		return nil, fmt.Errorf("central scheduler is required")
	}

	chainDVS, err := utils.NewDVSFromCfg(cfg, logger)
	if err != nil {
		logger.Error("failed to create chainDVS",
			"err", err,
			"file", pellcfg.CmtConfig.Pell.InteractorConfigPath,
			"cfg", fmt.Sprintf("%+v", cfg),
		)
		return nil, err
	}

	receipt, err := chainDVS.AddSuportedChain(ctx, approverPK, dvsChainID)
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

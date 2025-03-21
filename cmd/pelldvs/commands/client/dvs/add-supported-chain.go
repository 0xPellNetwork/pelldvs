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
	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

func init() {
	chainflags.DVSCentralSchedulerAddress.AddToCmdFlag(addSupportedChainCmd)
	chainflags.DVSEjectionManagerAddress.AddToCmdFlag(addSupportedChainCmd)
	chainflags.DVSOperatorStakeManagerAddress.AddToCmdFlag(addSupportedChainCmd)
	chainflags.DVSRPCURL.AddToCmdFlag(addSupportedChainCmd)
	chainflags.DVSApproverKeyName.AddToCmdFlag(addSupportedChainCmd)

	err := chainflags.MarkFlagsAreRequired(addSupportedChainCmd,
		chainflags.DVSCentralSchedulerAddress,
		chainflags.DVSEjectionManagerAddress,
		chainflags.DVSOperatorStakeManagerAddress,
		chainflags.DVSRPCURL,
		chainflags.DVSApproverKeyName,
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
	logger := getCmdLogger(cmd)
	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	appRoverKeyPath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.DVSApproverKeyName.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", appRoverKeyPath.ECDSA)
	}

	approverPkPair, err := ecdsa.ReadKey(appRoverKeyPath.ECDSA, "")
	if err != nil {
		return fmt.Errorf("failed to read approverPkPair ecdsa key: %v", err)
	}

	dvsETHClient, err := eth.NewClient(chainflags.DVSRPCURL.Value)
	if err != nil {
		return fmt.Errorf("failed to create eth client for RPCURL: %s:, err %v",
			chainflags.DVSRPCURL.Value, err,
		)
	}
	dvsChainID, err := dvsETHClient.ChainID(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get chain id from rpc client: url:%s , error: %v",
			chainflags.DVSRPCURL.Value, err)
	}

	receipt, err := execAddSupportedChain(cmd, logger, dvsChainID.Uint64(), kpath.ECDSA, approverPkPair)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execAddSupportedChain(cmd *cobra.Command, logger log.Logger, dvsChainID uint64, privKeyPath string, approverPK *osecdsa.PrivateKey) (*gethtypes.Receipt, error) {
	logger.Info("exec start",
		"privKeyPath", privKeyPath,
		"chainId", dvsChainID,
	)

	ctx := context.Background()
	senderAddress, err := ecdsa.GetAddressFromKeyStoreFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get senderAddress from key store file: %v", err)
	}

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

	if chainflags.DVSCentralSchedulerAddress.Value != "" {
		cfg.ContractConfig.DVSConfigs[dvsChainID].CentralScheduler = chainflags.DVSCentralSchedulerAddress.Value
	}
	if chainflags.DVSEjectionManagerAddress.Value != "" {
		cfg.ContractConfig.DVSConfigs[dvsChainID].EjectionManager = chainflags.DVSEjectionManagerAddress.Value
	}
	if chainflags.DVSOperatorStakeManagerAddress.Value != "" {
		cfg.ContractConfig.DVSConfigs[dvsChainID].OperatorStakeManager = chainflags.DVSOperatorStakeManagerAddress.Value
	}
	if chainflags.DVSRPCURL.Value != "" {
		cfg.ContractConfig.DVSConfigs[dvsChainID].RPCURL = chainflags.DVSRPCURL.Value
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
		"exec down",
		"k", "v",
		"senderAddress", senderAddress,
		"receipt", receipt,
	)

	return receipt, err
}

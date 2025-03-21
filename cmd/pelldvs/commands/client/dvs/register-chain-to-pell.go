package dvs

import (
	"context"
	ecdsa2 "crypto/ecdsa"
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
	chainflags.DVSCentralSchedulerAddress.AddToCmdFlag(registerChainToPellCmd)
	chainflags.DVSRPCURL.AddToCmdFlag(registerChainToPellCmd)
	chainflags.DVSApproverKeyName.AddToCmdFlag(registerChainToPellCmd)
	chainflags.DVSFrom.AddToCmdFlag(registerChainToPellCmd)

	err := chainflags.MarkFlagsAreRequired(registerChainToPellCmd,
		chainflags.PellRegistryRouterAddress,
		chainflags.DVSCentralSchedulerAddress,
		chainflags.DVSApproverKeyName,
		chainflags.DVSFrom,
	)
	if err != nil {
		panic(err)
	}
}

var registerChainToPellCmd = &cobra.Command{
	Use:   "register-chain-to-pell",
	Short: "register-chain-to-pell",
	Long: `
pelldvs client dvs register-chain-to-pell \
	--rpc-url <rpc-url> \
	--registry-router <registry-router-address> \
	--central-scheduler <central-scheduler> \
	--dvs-rpc-url <dvs-rpc-url> \
	--dvs-from <dvs-from> \
	--approver-key-name <approver-key-name>
`,
	Example: `
pelldvs client dvs register-chain-to-pell \
	--rpc-url http://localhost:8646 \
	--registry-router "0xE7402c51ae0bd667ad57a99991af6C2b686cd4f1" \
	--central-scheduler "0xdbC43Ba45381e02825b14322cDdd15eC4B3164E6" \
	--dvs-rpc-url http://localhost:8747 \
	--dvs-from pell-localnet-deployer \
	--approver-key-name pell-localnet-deployer

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleRegisterChainToPell(cmd)
	},
}

func handleRegisterChainToPell(cmd *cobra.Command) error {
	logger := getCmdLogger(cmd)
	approverKeyPath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.DVSApproverKeyName.Value)
	if !approverKeyPath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", approverKeyPath.ECDSA)
	}

	dvsFromKeyPath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.DVSFrom.Value)
	if !dvsFromKeyPath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", dvsFromKeyPath.ECDSA)
	}

	approverPkPair, err := ecdsa.ReadKey(approverKeyPath.ECDSA, "")
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

	receipt, err := execRegisterChainToPell(cmd, logger, dvsChainID.Uint64(), dvsFromKeyPath.ECDSA, approverPkPair)

	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}
	logger.Info("tx successfully", "txHash", receipt.TxHash.String())
	return err
}

func execRegisterChainToPell(cmd *cobra.Command,
	logger log.Logger,
	chainID uint64,
	dvsFromKeyPath string,
	appriverPk *ecdsa2.PrivateKey,
) (*gethtypes.Receipt, error) {
	cmdName := utils.GetPrettyCommandName(cmd)
	logger.Info(cmdName,
		"chainID", chainID,
	)

	ctx := context.Background()
	cfg, err := utils.LoadChainConfig(cmd, pellcfg.CmtConfig.Pell.InteractorConfigPath, logger)
	if err != nil {
		logger.Error("failed to load chain config", "err", err, "file", pellcfg.CmtConfig.Pell.InteractorConfigPath)
		return nil, err
	}

	if cfg.ContractConfig.DVSConfigs == nil {
		cfg.ContractConfig.DVSConfigs = make(map[uint64]*chaincfg.DVSConfig)
	}
	if cfg.ContractConfig.DVSConfigs[chainID] == nil {
		cfg.ContractConfig.DVSConfigs[chainID] = &chaincfg.DVSConfig{}
	}

	if chainflags.DVSCentralSchedulerAddress.Value != "" {
		cfg.ContractConfig.DVSConfigs[chainID].CentralScheduler = chainflags.DVSCentralSchedulerAddress.Value
	}
	if chainflags.DVSRPCURL.Value != "" {
		cfg.ContractConfig.DVSConfigs[chainID].RPCURL = chainflags.DVSRPCURL.Value
	}

	cfg.ContractConfig.DVSConfigs[chainID].ECDSAPrivateKeyFilePath = dvsFromKeyPath

	logger.Info("chain config details", "chaincfg", fmt.Sprintf("%+v", cfg))

	var chainConfigChecker = utils.NewChainConfigChecker(cfg)
	if !chainConfigChecker.HasRPCURL() {
		return nil, fmt.Errorf("rpc url is required")
	}
	if !chainConfigChecker.IsValidPellRegistryRouter() {
		return nil, fmt.Errorf("pell registry router is required")
	}
	if !chainConfigChecker.IsValidDVSCentralScheduler(chainID) {
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

	receipt, err := chainDVS.RegisterChainToPell(ctx, appriverPk, chainID)

	logger.Info(cmdName,
		"k", "v",
		"receipt", receipt,
	)

	return receipt, err
}

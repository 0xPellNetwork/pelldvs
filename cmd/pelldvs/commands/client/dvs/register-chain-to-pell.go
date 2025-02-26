package dvs

import (
	"context"
	ecdsa2 "crypto/ecdsa"
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

// TODO(jimmy): remove --from flag, it is not needed for this command
// TODO(jimmy): remove --chain-id flag, it's can be get from DVSRPCClient

var (
	registerChainToPellCmdFlagCentralScheduler = &chainflags.StringFlag{
		Name:  "central-scheduler",
		Usage: "central scheduler address",
	}
	registerChainToPellCmdFlagDVSRPCURL = &chainflags.StringFlag{
		Name:  "dvs-rpc-url",
		Usage: "dvs rpc url",
	}
	registerChainToPellCmdFlagDVSFrom = &chainflags.StringFlag{
		Name:  "dvs-from",
		Usage: "dvs from key name",
	}
	registerChainToPellCmdFlagApproverKeyName = &chainflags.StringFlag{
		Name:  "approver-key-name",
		Usage: "approver key name",
	}
)

func init() {
	chainflags.ChainIDFlag.AddToCmdFlag(registerChainToPellCmd)
	registerChainToPellCmdFlagCentralScheduler.AddToCmdFlag(registerChainToPellCmd)
	registerChainToPellCmdFlagDVSRPCURL.AddToCmdFlag(registerChainToPellCmd)
	registerChainToPellCmdFlagApproverKeyName.AddToCmdFlag(registerChainToPellCmd)
	registerChainToPellCmdFlagDVSFrom.AddToCmdFlag(registerChainToPellCmd)

	err := chainflags.MarkFlagsAreRequired(registerChainToPellCmd,
		chainflags.ChainIDFlag,
		chainflags.PellRegistryRouterAddress,
		registerChainToPellCmdFlagCentralScheduler,
		registerChainToPellCmdFlagApproverKeyName,
		registerChainToPellCmdFlagDVSFrom,
	)
	if err != nil {
		panic(err)
	}
}

var registerChainToPellCmd = &cobra.Command{
	Use:   "register-chain-to-pell",
	Short: "register-chain-to-pell",
	Long:  `register-chain-to-pell`,
	Example: `

pelldvs client dvs register-chain-to-pell \
	--from <key-name> \
	--rpc-url <rpc-url> \
	--registry-router <registry-router-address> \
	--chain-id <chain-id> \
	--central-scheduler <central-scheduler> \
	--dvs-rpc-url <dvs-rpc-url> \
	--dvs-from <dvs-from> \
	--approver-key-name <approver-key-name>

pelldvs client dvs register-chain-to-pell \
	--from pell-localnet-deployer \
	--rpc-url http://localhost:8646 \
	--registry-router "0xE7402c51ae0bd667ad57a99991af6C2b686cd4f1" \
	--chain-id 1337 \
	--central-scheduler "0xdbC43Ba45381e02825b14322cDdd15eC4B3164E6" \
	--dvs-rpc-url http://localhost:8747 \
	--dvs-from pell-localnet-deployer \
	--approver-key-name pell-localnet-deployer

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleRegisterChainToPell(
			cmd,
			uint64(chainflags.ChainIDFlag.Value),
			registerChainToPellCmdFlagDVSFrom.Value,
			registerChainToPellCmdFlagApproverKeyName.Value,
		)
	},
}

func handleRegisterChainToPell(cmd *cobra.Command, chainID uint64, DVSFrom, approverKeyName string) error {

	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	approverKeyPath := keys.GetKeysPath(pellcfg.CmtConfig, approverKeyName)
	if !approverKeyPath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", approverKeyPath.ECDSA)
	}

	dvsFromKeyPath := keys.GetKeysPath(pellcfg.CmtConfig, DVSFrom)
	if !dvsFromKeyPath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", dvsFromKeyPath.ECDSA)
	}

	approverPkPair, err := ecdsa.ReadKey(approverKeyPath.ECDSA, "")
	if err != nil {
		return fmt.Errorf("failed to read approverPkPair ecdsa key: %v", err)
	}

	receipt, err := execRegisterChainToPell(cmd, kpath.ECDSA, chainID, dvsFromKeyPath.ECDSA, approverPkPair)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}
	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execRegisterChainToPell(cmd *cobra.Command,
	privKeyPath string,
	chainID uint64,
	dvsFromKeyPath string,
	appriverPk *ecdsa2.PrivateKey,
) (*gethtypes.Receipt, error) {
	cmdName := "handleRegisterChainToPell"

	logger.Info(cmdName,
		"privKeyPath", privKeyPath,
		"chainID", chainID,
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

	if registerChainToPellCmdFlagCentralScheduler.Value != "" {
		cfg.ContractConfig.DVSConfigs[chainID].CentralScheduler = registerChainToPellCmdFlagCentralScheduler.Value
	}
	if registerChainToPellCmdFlagDVSRPCURL.Value != "" {
		cfg.ContractConfig.DVSConfigs[chainID].RPCURL = registerChainToPellCmdFlagDVSRPCURL.Value
	}

	cfg.ContractConfig.DVSConfigs[chainID].ECDSAPrivateKeyFilePath = dvsFromKeyPath

	logger.Debug("chaincfg", "cfg", fmt.Sprintf("%+v", cfg))

	chainDVS, err := utils.NewDVSFromCfg(cfg, logger)
	if err != nil {
		return nil, err
	}

	receipt, err := chainDVS.RegisterChainToPell(ctx, appriverPk, chainID)

	logger.Info(cmdName,
		"k", "v",
		"sender", senderAddress,
		"receipt", receipt,
	)

	return receipt, err
}

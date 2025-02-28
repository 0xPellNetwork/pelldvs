package dvs

import (
	"context"
	"fmt"

	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

var setEjectorCmd = &cobra.Command{
	Use:   "set-ejector",
	Short: "set-ejector",
	Long: `
  /**
   * @notice Sets the ejector, which can force-deregister operators from groups
   * @param _ejector the new ejector
   * @dev only callable by the owner
   */

pelldvs client dvs set-ejector \
	--from <key-name> \
	--rpc-url <rpc-url> \
	--registry-router <registry-router> \
	<ejector-address>

`,
	Args: cobra.ExactArgs(1),
	Example: `
pelldvs client dvs set-ejector \
	--from pell-localnet-deployer \
	--rpc-url http://127.0.0.1:8545 \
	--registry-router 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f \
0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleSetEjector(cmd, args[0])
	},
}

func handleSetEjector(cmd *cobra.Command, ejector string) error {
	if !gethcommon.IsHexAddress(ejector) {
		return fmt.Errorf("invalid address %s", ejector)
	}

	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	receipt, err := execSetEjector(cmd, kpath.ECDSA, ejector)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execSetEjector(cmd *cobra.Command, privKeyPath string, ejector string) (*gethtypes.Receipt, error) {
	cmdName := utils.GetPrettyCommandName(cmd)
	logger.Info(cmdName,
		"privKeyPath", privKeyPath,
		"ejector", ejector,
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
	logger.Info("chain config details", "chaincfg", fmt.Sprintf("%+v", cfg))

	var chainConfigChecker = utils.NewChainConfigChecker(cfg)
	if !chainConfigChecker.HasRPCURL() {
		return nil, fmt.Errorf("rpc url is required")
	}
	if !chainConfigChecker.IsValidPellRegistryRouter() {
		return nil, fmt.Errorf("pell registry router is required")
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

	receipt, err := chainDVS.SetEjector(ctx, ejector)

	logger.Info(cmdName,
		"k", "v",
		"sender", senderAddress,
		"receipt", receipt,
	)

	return receipt, err
}

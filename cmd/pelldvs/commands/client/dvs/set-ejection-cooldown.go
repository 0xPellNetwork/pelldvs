package dvs

import (
	"context"
	"fmt"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainutils"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

var setEjectionCooldownCmd = &cobra.Command{
	Use:   "set-ejection-cooldown",
	Short: "set-ejection-cooldown",
	Long: `Sets the ejection cooldown, which is the time an operator must wait in
   * seconds afer ejection before registering for any group
   * @param _ejectionCooldown the new ejection cooldown in seconds

pelldvs client dvs set-ejection-cooldown\
	--from <key-name> \
	--rpc-url <rpc-url> \
	--registry-router <registry-router> \
	<seconds>
`,
	Args: cobra.ExactArgs(1),
	Example: `
pelldvs client dvs set-ejection-cooldown \
	--from pell-localnet-deployer \
	--rpc-url http://127.0.0.1:8545 \
	--registry-router 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f \
	666

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cooldown, err := chainutils.ConvStrToUint64(args[0])
		if err != nil {
			return fmt.Errorf("failed to convert `%s` to Uint64, cause: %v", args[0], err)
		}

		return handleSetEjectionCooldown(cmd, cooldown)
	},
}

func handleSetEjectionCooldown(cmd *cobra.Command, cooldown uint64) error {
	logger := getCmdLogger(cmd)
	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	receipt, err := execSetEjectionCooldown(cmd, logger, kpath.ECDSA, cooldown)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execSetEjectionCooldown(cmd *cobra.Command, logger log.Logger, privKeyPath string, cooldown uint64) (*gethtypes.Receipt, error) {
	ctx := context.Background()
	address, err := ecdsa.GetAddressFromKeyStoreFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get address from key store file: %v", err)
	}

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

	receipt, err := chainDVS.SetEjectionCooldown(ctx, cooldown)

	logger.Info(
		utils.GetPrettyCommandName(cmd),
		"k", "v",
		"sender", address,
		"receipt", receipt,
	)

	return receipt, err
}

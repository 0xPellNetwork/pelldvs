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

var setMinimumStakeForGroupCmd = &cobra.Command{
	Use:   "set-minimum-stake-for-group",
	Short: "set-minimum-stake-for-group",
	Args:  cobra.ExactArgs(2),
	Long: `
pelldvs client dvs set-minimum-stake-for-group \
	--from <key-name> \
	--rpc-url <rpc-url> \
	--registry-router <registry-router> \
	<number> <stake>
`,
	Example: `
pelldvs client dvs set-minimum-stake-for-group \
	--from pell-localnet-deployer \
	--rpc-url http://127.0.0.1:8545 \
	--registry-router 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f \
	1 1000000000000000000
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		groupNumber, err := chainutils.ConvStrToUint8(args[0])
		if err != nil {
			return fmt.Errorf("failed to convert to uint8: %v", err)
		}
		minimumStake, err := chainutils.ConvStrToUint64(args[1])
		if err != nil {
			return fmt.Errorf("can't convert `%s` to Uint64, cause: %v ", args[1], err)
		}

		return handleSetMinimumStakeForGroup(cmd, groupNumber, minimumStake)
	},
}

func handleSetMinimumStakeForGroup(cmd *cobra.Command, groupNumber uint8, minimumStake uint64) error {
	logger := getCmdLogger(cmd)
	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	receipt, err := execSetMinimumStakeForGroup(cmd, logger, kpath.ECDSA, groupNumber, minimumStake)
	if err != nil {
		return fmt.Errorf("failed to handleSetMinimumStakeForGroup: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execSetMinimumStakeForGroup(cmd *cobra.Command, logger log.Logger, privKeyPath string, groupNumber uint8, minimumStake uint64) (*gethtypes.Receipt, error) {
	logger.Info(
		utils.GetPrettyCommandName(cmd),
		"groupNumber", groupNumber,
		"minimumStake", minimumStake,
	)

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

	receipt, err := chainDVS.SetMinimumStakeForGroup(ctx, groupNumber, minimumStake)

	logger.Info(
		utils.GetPrettyCommandName(cmd),
		"k", "v",
		"sender", address,
		"receipt", receipt,
	)

	return receipt, err
}

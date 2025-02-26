package dvs

import (
	"context"
	"fmt"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
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
	Long:  ``,
	Example: `
pelldvs client dvs set-minimum-stake-for-group --from <key-name> <number> <stake>
pelldvs client dvs set-minimum-stake-for-group --from pell-localnet-deployer 1 1000000000000000000

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
	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	receipt, err := execSetMinimumStakeForGroup(cmd, kpath.ECDSA, groupNumber, minimumStake)
	if err != nil {
		return fmt.Errorf("failed to handleSetMinimumStakeForGroup: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execSetMinimumStakeForGroup(cmd *cobra.Command, privKeyPath string, groupNumber uint8, minimumStake uint64) (*gethtypes.Receipt, error) {
	logger.Info(
		"execSetMinimumStakeForGroup",
		"groupNumber", groupNumber,
		"minimumStake", minimumStake,
		"ethRPCURL", chainflags.EthRPCURLFlag.Value,
	)

	ctx := context.Background()
	address, err := ecdsa.GetAddressFromKeyStoreFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get address from key store file: %v", err)
	}

	chainDVS, _, err := utils.NewDVSFromFromFile(cmd, pellcfg.CmtConfig.Pell.InteractorConfigPath, logger)
	if err != nil {
		logger.Error("failed to create operator", "err", err, "file", pellcfg.CmtConfig.Pell.InteractorConfigPath)
		return nil, err
	}

	receipt, err := chainDVS.SetMinimumStakeForGroup(ctx, groupNumber, minimumStake)

	logger.Info(
		"execSetMinimumStakeForGroup",
		"k", "v",
		"sender", address,
		"receipt", receipt,
	)

	return receipt, err
}

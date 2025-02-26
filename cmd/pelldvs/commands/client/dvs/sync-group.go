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

func init() {
	chainflags.ChainIDFlag.AddToCmdFlag(syncGroupCmd)
	chainflags.GroupNumbers.AddToCmdFlag(syncGroupCmd)
	err := chainflags.MarkFlagsAreRequired(syncGroupCmd, chainflags.ChainIDFlag)
	if err != nil {
		panic(err)
	}
}

var syncGroupCmd = &cobra.Command{
	Use:   "sync-group",
	Short: "sync-group",
	Long:  `sync-group`,
	Example: `

pelldvs client dvs sync-group --from pell-localnet-deployer
pelldvs client dvs sync-group --from pell-localnet-deployer

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if chainflags.ChainIDFlag.Value == 0 {
			chainflags.ChainIDFlag.Value = 1337
		}

		if chainflags.GroupNumbers.Value == "" {
			chainflags.GroupNumbers.Value = "0"
		}

		return handleSyncGroup(cmd, chainflags.FromKeyNameFlag.Value, chainflags.ChainIDFlag.Value, chainflags.GroupNumbers.Value)
	},
}

func handleSyncGroup(cmd *cobra.Command, keyName string, chainID int, groupNumbersStr string) error {
	groupNumbers := chainutils.ConvStrsToUint8List(groupNumbersStr)
	if len(groupNumbers) == 0 {
		return fmt.Errorf("invalid group numbers %s", groupNumbersStr)
	}

	kpath := keys.GetKeysPath(pellcfg.CmtConfig, keyName)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	res, err := execSyncGroup(cmd, kpath.ECDSA, chainID, groupNumbers)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}
	logger.Info("tx successfully", "txHash ", res.TxHash.String())

	return err
}

func execSyncGroup(cmd *cobra.Command, privKeyPath string, chainID int, groupNumbers []byte) (*gethtypes.Receipt, error) {
	cmdName := "handleSyncGroup"

	logger.Info(cmdName,
		"privKeyPath", privKeyPath,
	)

	ctx := context.Background()
	senderAddress, err := ecdsa.GetAddressFromKeyStoreFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get senderAddress from key store file: %v", err)
	}
	logger.Info(cmdName,
		"sender", senderAddress,
	)

	chainDVS, _, err := utils.NewDVSFromFromFile(cmd, pellcfg.CmtConfig.Pell.InteractorConfigPath, logger)
	if err != nil {
		logger.Error("failed to create operator", "err", err, "file", pellcfg.CmtConfig.Pell.InteractorConfigPath)
		return nil, err
	}
	receipt, err := chainDVS.SyncGroup(ctx, uint64(chainID), groupNumbers)

	logger.Info(cmdName,
		"k", "v",
		"sender", senderAddress,
		"receipt", receipt,
	)

	return receipt, err
}

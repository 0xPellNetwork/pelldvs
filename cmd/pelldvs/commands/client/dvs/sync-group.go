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
	Long: `
pelldvs client dvs sync-group \
	--from <key-name> \
	--rpc-url <rpc-url> \
	--registry-router <registry-router> \
	--chain-id <chain-id, defaults to 1337> \
	<group-numbers, defaults to 0>
`,
	Example: `
pelldvs client dvs sync-group \
	--from pell-localnet-deployer
	--rpc-url http://localhost:8545 \
	--registry-router 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f \
	--chain-id 666 \
	0
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleSyncGroup(cmd)
	},
}

func handleSyncGroup(cmd *cobra.Command) error {
	groupNumbersStr := chainflags.GroupNumbers.Value
	if groupNumbersStr == "" {
		groupNumbersStr = "0"
	}
	if chainflags.ChainIDFlag.Value == 0 {
		chainflags.ChainIDFlag.Value = 1337
	}

	groupNumbers := chainutils.ConvStrsToUint8List(groupNumbersStr)
	if len(groupNumbers) == 0 {
		return fmt.Errorf("invalid group numbers `%s`", groupNumbersStr)
	}

	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	res, err := execSyncGroup(cmd, kpath.ECDSA, chainflags.ChainIDFlag.Value, groupNumbers)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}
	logger.Info("tx successfully", "txHash ", res.TxHash.String())

	return err
}

func execSyncGroup(cmd *cobra.Command, privKeyPath string, chainID int, groupNumbers []byte) (*gethtypes.Receipt, error) {
	cmdName := utils.GetPrettyCommandName(cmd)
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

	receipt, err := chainDVS.SyncGroup(ctx, uint64(chainID), groupNumbers)

	logger.Info(cmdName,
		"k", "v",
		"sender", senderAddress,
		"receipt", receipt,
	)

	return receipt, err
}

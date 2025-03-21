package operator

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

var deRegisterOperatorCmd = &cobra.Command{
	Use:   "deregister-operator",
	Short: "deregister-operator",
	Args:  cobra.ExactArgs(1),
	Long: `Deregisters the caller from one or more groups
* @param groupNumbers is an ordered byte array containing the group numbers being deregistered from

pelldvs client operator deregister-operator \ 
	--from <key-name> \
	--rpc-url <rpc-url> \
	--registry-router <registry-router> \
	<group-numbers>
`,
	Example: `

pelldvs client operator deregister-operator \
	--from pell-localnet-deployer
	--rpc-url http://localhost:8545 \
	--registry-router 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f \
	1,2,3
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleDeRegisterOperator(cmd, args[0])
	},
}

func handleDeRegisterOperator(cmd *cobra.Command, groupNumbersStr string) error {
	logger := getCmdLogger(cmd)
	groupNumbers := chainutils.ConvStrsToUint8List(groupNumbersStr)
	if len(groupNumbers) == 0 {
		return fmt.Errorf("invalid group numbers %s", groupNumbersStr)
	}
	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	receipt, err := execDeRegisterOperator(cmd, logger, kpath.ECDSA, groupNumbers)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execDeRegisterOperator(cmd *cobra.Command, logger log.Logger, privKeyPath string, groupNumbers []byte) (*gethtypes.Receipt, error) {
	logger.Info(
		utils.GetPrettyCommandName(cmd),
		"privKeyPath", privKeyPath,
		"operator", groupNumbers,
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

	chainOp, err := utils.NewOperatorFromCfg(cfg, logger)
	if err != nil {
		logger.Error("failed to create chain operator",
			"err", err,
			"file", pellcfg.CmtConfig.Pell.InteractorConfigPath,
			"cfg", fmt.Sprintf("%+v", cfg),
		)
		return nil, err
	}

	receipt, err := chainOp.DeregisterOperator(ctx, groupNumbers)

	logger.Info(
		utils.GetPrettyCommandName(cmd),
		"address", address,
		"receipt", receipt,
	)

	return receipt, err
}

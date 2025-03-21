package dvs

import (
	"context"
	"fmt"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	chaintypes "github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	cmtos "github.com/0xPellNetwork/pelldvs/libs/os"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

var setOperatorSetCmd = &cobra.Command{
	Use:   "set-operator-set-param",
	Short: "set operator set param",
	Args:  cobra.ExactArgs(1),
	Long: `  /**
   * @notice Updates an existing group's configuration with a new max operator count
   * and operator churn parameters
   * @param groupNumber the group number to update
   * @param operatorSetParams the new config
   * @dev only callable by the owner
   */


pelldvs client dvs set-operator-set-params \
	--from <key-name> \
	--rpc-url <rpc-url> \
	--registry-router <registry-router> \
	<param-file-path.json>

`,
	Example: `
pelldvs client dvs set-operator-set-param \
	--from pell-localnet-deployer \
	--rpc-url https://localhost:8545 \
	--registry-router 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f \
	/data/pells/dvsreqs2/set-operator-set-param-1.json
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		paramFilePath := args[0]
		return handleSetOperatorSet(cmd, paramFilePath)
	},
}

func handleSetOperatorSet(cmd *cobra.Command, paramFilePath string) error {
	logger := getCmdLogger(cmd)
	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	// TODO(jimmy):  password support
	if !cmtos.FileExists(paramFilePath) {
		return fmt.Errorf("param file does not exist %s", paramFilePath)
	}

	var setOperatorSetParam chaintypes.SetOperatorSetParamRequest
	err := decodeJSONFromFile(paramFilePath, &setOperatorSetParam)
	if err != nil {
		return fmt.Errorf("failed to unmarshal setOperatorSetParam: %v", err)
	}

	receipt, err := execSetOperatorSet(cmd, logger, &setOperatorSetParam, kpath.ECDSA)
	if err != nil {
		return fmt.Errorf("failed to handleSetOperatorSet: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execSetOperatorSet(cmd *cobra.Command, logger log.Logger, params *chaintypes.SetOperatorSetParamRequest, privKeyPath string) (*gethtypes.Receipt, error) {
	cmdName := utils.GetPrettyCommandName(cmd)
	logger.Info(fmt.Sprintf("%s start", cmdName),
		"privKeyPath", privKeyPath,
		"params", params,
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

	receipt, err := chainDVS.SetOperatorSetParams(ctx, params)
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("%s done", cmdName),
		"k", "v",
		"sender", senderAddress,
		"receipt", receipt,
	)

	return receipt, err
}

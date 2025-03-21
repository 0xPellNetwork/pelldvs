package operator

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
)

var isOperatorCmd = &cobra.Command{
	Use:   "is-operator",
	Short: "is-operator",
	Args:  cobra.ExactArgs(1),
	Long: `
  /**
   * @notice Returns true is an operator has previously registered for delegation.
   */

pelldvs client operator is-operator \
	--rpc-url <rpc-url> \
	--delegation-manager <delegation-manager> \
	<operator-address>

`,
	Example: `
pelldvs client operator is-operator \
	--rpc-url http://localhost:8545 \
	--delegation-manager 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f \
	0xa0Ee7A142d267C1f36714E4a8F75612F20a79720
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleCheckIsOperator(cmd, args[0])
	},
}

func handleCheckIsOperator(cmd *cobra.Command, operatorAddr string) error {
	logger := getCmdLogger(cmd)
	if !gethcommon.IsHexAddress(operatorAddr) {
		return fmt.Errorf("invalid address %s", operatorAddr)
	}
	result, err := execCheckIsOperator(cmd, logger, operatorAddr)
	if err != nil {
		return fmt.Errorf("failed to execCheckIsOperator: %v", err)
	}

	logger.Info("tx successfully", "is-operator", result)

	return err
}

func execCheckIsOperator(cmd *cobra.Command, logger log.Logger, operatorAddr string) (bool, error) {
	logger.Info(
		utils.GetPrettyCommandName(cmd),
		"operatorAddr", operatorAddr,
	)

	cfg, err := utils.LoadChainConfig(cmd, pellcfg.CmtConfig.Pell.InteractorConfigPath, logger)
	if err != nil {
		logger.Error("failed to load chain config", "err", err, "file", pellcfg.CmtConfig.Pell.InteractorConfigPath)
		return false, err
	}

	logger.Info("chain config details", "chaincfg", fmt.Sprintf("%+v", cfg))

	var chainConfigChecker = utils.NewChainConfigChecker(cfg)
	if !chainConfigChecker.HasRPCURL() {
		return false, fmt.Errorf("rpc url is required")
	}
	if !chainConfigChecker.IsValidPellDelegationManager() {
		return false, fmt.Errorf("pell delegation manager is required")
	}

	chainOp, err := utils.NewOperatorFromCfg(cfg, logger)
	if err != nil {
		logger.Error("failed to create chain operator",
			"err", err,
			"file", pellcfg.CmtConfig.Pell.InteractorConfigPath,
			"cfg", fmt.Sprintf("%+v", cfg),
		)
		return false, err
	}

	ctx := context.Background()
	result, err := chainOp.IsOperator(&bind.CallOpts{Context: ctx}, operatorAddr)

	logger.Info(
		utils.GetPrettyCommandName(cmd),
		"address", operatorAddr,
		"result", result,
	)

	return result, err
}

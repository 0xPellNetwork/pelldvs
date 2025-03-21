package operator

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
)

var operatorDetailsCmd = &cobra.Command{
	Use:   "operator-details",
	Short: "operator-details",
	Args:  cobra.ExactArgs(1),
	Long: `
  /**
   * @notice Returns the OperatorDetails struct associated with an operator.
   */

pelldvs query operator operator-details \
	--rpc-url <rpc-url> \
	--delegation-manager <delegation-manager> \
	<operator-address>
`,
	Example: `
pelldvs query operator operator-details \
	--rpc-url http://localhost:8545 \
	--delegation-manager 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f \
0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleQueryOperatorDetails(cmd, args[0])
	},
}

func handleQueryOperatorDetails(cmd *cobra.Command, operatorAddr string,
) error {
	logger := getCmdLogger(cmd)
	if !gethcommon.IsHexAddress(operatorAddr) {
		return fmt.Errorf("invalid address %s", operatorAddr)
	}
	var err error

	result, err := execQueryOperatorDetails(cmd, logger, operatorAddr)
	if err != nil {
		return fmt.Errorf("failed to create group: %v", err)
	}

	logger.Info("tx successfully", "result", result)

	return err
}

func execQueryOperatorDetails(cmd *cobra.Command, logger log.Logger, operator string) (*types.Operator, error) {
	logger.Info(
		utils.GetPrettyCommandName(cmd),
		"operator", operator,
	)

	ctx := context.Background()
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
	if !chainConfigChecker.IsValidPellDelegationManager() {
		return nil, fmt.Errorf("pell delegation manager is required")
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

	result, err := chainOp.GetOperatorDetails(&bind.CallOpts{Context: ctx}, operator)

	logger.Info(
		utils.GetPrettyCommandName(cmd),
		"result", result,
	)

	return result, err
}

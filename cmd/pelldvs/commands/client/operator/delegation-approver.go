package operator

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
)

var delegationApproverCmd = &cobra.Command{
	Use:   "delegation-approver",
	Short: "delegation-approver",
	Args:  cobra.ExactArgs(1),
	Long: `
  /**
   * @notice Returns the handleDelegationApprover account for an operator
   */

pelldvs query operator delegation-approver \
	--rpc-url <rpc-url> \
	--delegation-manager <delegation-manager> \
	<operator-address>

`,
	Example: `
pelldvs query operator delegation-approver \
	--rpc-url http://localhost:8545 \
	--delegation-manager 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f \
	0xa0Ee7A142d267C1f36714E4a8F75612F20a79720

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleDelegationApprover(cmd, args[0])
	},
}

func handleDelegationApprover(cmd *cobra.Command, operatorAddr string) error {
	logger := getCmdLogger(cmd)
	result, err := execDelegationApprover(cmd, logger, operatorAddr)
	if err != nil {
		return fmt.Errorf("failed to create group: %v", err)
	}

	result = strings.TrimSpace(result)

	var found = true
	if result == "0x0000000000000000000000000000000000000000" || result == "" {
		found = false
	}

	logger.Info("tx successfully", "found", found, "result", result)

	return err
}

func execDelegationApprover(cmd *cobra.Command, logger log.Logger, operator string) (string, error) {
	logger.Info(
		utils.GetPrettyCommandName(cmd),
		"operator", operator,
	)

	ctx := context.Background()
	cfg, err := utils.LoadChainConfig(cmd, pellcfg.CmtConfig.Pell.InteractorConfigPath, logger)
	if err != nil {
		logger.Error("failed to load chain config", "err", err, "file", pellcfg.CmtConfig.Pell.InteractorConfigPath)
		return "", err
	}

	logger.Info("chain config details", "chaincfg", fmt.Sprintf("%+v", cfg))

	var chainConfigChecker = utils.NewChainConfigChecker(cfg)
	if !chainConfigChecker.HasRPCURL() {
		return "", fmt.Errorf("rpc url is required")
	}
	if !chainConfigChecker.IsValidPellDelegationManager() {
		return "", fmt.Errorf("pell delegation manager is required")
	}

	chainOp, err := utils.NewOperatorFromCfg(cfg, logger)
	if err != nil {
		logger.Error("failed to create chain operator",
			"err", err,
			"file", pellcfg.CmtConfig.Pell.InteractorConfigPath,
			"cfg", fmt.Sprintf("%+v", cfg),
		)
		return "", err
	}

	result, err := chainOp.GetDelegationApprover(&bind.CallOpts{Context: ctx}, operator)

	logger.Info(
		utils.GetPrettyCommandName(cmd),
		"result", result,
	)

	return result, err
}

package operator

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

var delegationApproverCmd = &cobra.Command{
	Use:   "delegation-approver",
	Short: "delegation-approver",
	Args:  cobra.ExactArgs(1),
	Long: `
  /**
   * @notice Returns the handleDelegationApprover account for an operator
   */
`,
	Example: `

pelldvs query operator delegation-approver --from <key-name> <operator-address>

pelldvs query operator delegation-approver --from pell-localnet-deployer 0xa0Ee7A142d267C1f36714E4a8F75612F20a79720
pelldvs query operator delegation-approver --from pell-localnet-deployer 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleDelegationApprover(cmd, chainflags.FromKeyNameFlag.Value, args[0])
	},
}

func handleDelegationApprover(cmd *cobra.Command, keyName string,
	operatorAddr string,
) error {

	kpath := keys.GetKeysPath(pellcfg.CmtConfig, keyName)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	result, err := execDelegationApprover(cmd, kpath.ECDSA, operatorAddr)
	if err != nil {
		return fmt.Errorf("failed to create group: %v", err)
	}

	var res = result
	if result == "0x0000000000000000000000000000000000000000" {
		res = "0x ---- not found"
	}

	logger.Info("tx successfully", "result", res)

	return err
}

func execDelegationApprover(cmd *cobra.Command, privKeyPath string, operator string) (string, error) {
	logger.Info(
		"handleDelegationApprover",
		"ethRPCURL", chainflags.EthRPCURLFlag.Value,
		"privKeyPath", privKeyPath,
		"operator", operator,
	)

	ctx := context.Background()
	address, err := ecdsa.GetAddressFromKeyStoreFile(privKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to get address from key store file: %v", err)
	}

	chainOp, _, err := utils.NewOperatorFromFile(cmd, pellcfg.CmtConfig.Pell.InteractorConfigPath, logger)
	if err != nil {
		logger.Error("failed to create operator", "err", err)
		return "", err
	}

	result, err := chainOp.GetDelegationApprover(&bind.CallOpts{Context: ctx}, operator)

	logger.Info(
		"handleDelegationApprover",
		"address", address,
		"result", result,
	)

	return result, err
}

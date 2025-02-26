package operator

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

var isOperatorCmd = &cobra.Command{
	Use:   "is-operator",
	Short: "is-operator",
	Args:  cobra.ExactArgs(1),
	Long: `
  /**
   * @notice Returns true is an operator has previously registered for delegation.
   */
`,
	Example: `
pelldvs client operator is-operator --from <key-name> <operator-address>

pelldvs client operator is-operator --from pell-localnet-deployer 0xa0Ee7A142d267C1f36714E4a8F75612F20a79720
pelldvs client operator is-operator --from pell-localnet-deployer 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleCheckIsOperator(cmd, chainflags.FromKeyNameFlag.Value, args[0])
	},
}

func handleCheckIsOperator(cmd *cobra.Command, keyName string,
	operatorAddr string,
) error {
	if !gethcommon.IsHexAddress(operatorAddr) {
		return fmt.Errorf("invalid address %s", operatorAddr)
	}
	var err error

	kpath := keys.GetKeysPath(pellcfg.CmtConfig, keyName)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	_, err = ecdsa.ReadKey(kpath.ECDSA, "")
	if err != nil {
		return fmt.Errorf("failed to read ecdsa key: %v", err)
	}

	result, err := execCheckIsOperator(cmd, kpath.ECDSA, operatorAddr)
	if err != nil {
		return fmt.Errorf("failed to execCheckIsOperator: %v", err)
	}

	logger.Info("tx successfully", "result", result)

	return err
}

func execCheckIsOperator(cmd *cobra.Command, privKeyPath string, operatorAddress string) (bool, error) {
	logger.Info(
		"handleCheckIsOperator",
		"ethRPCURL", chainflags.EthRPCURLFlag.Value,
		"privKeyPath", privKeyPath,
		"operatorAddress", operatorAddress,
	)

	chainOp, _, err := utils.NewOperatorFromFile(cmd, pellcfg.CmtConfig.Pell.InteractorConfigPath, logger)
	if err != nil {
		logger.Error("failed to create operator", "err", err)
		return false, err
	}

	ctx := context.Background()
	result, err := chainOp.IsOperator(&bind.CallOpts{Context: ctx}, operatorAddress)

	logger.Info(
		"handleCheckIsOperator",
		"address", operatorAddress,
		"result", result,
	)

	return result, err
}

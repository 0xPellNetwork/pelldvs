package operator

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

var operatorDetailsCmd = &cobra.Command{
	Use:   "operator-details",
	Short: "operator-details",
	Args:  cobra.ExactArgs(1),
	Long: `
  /**
   * @notice Returns the OperatorDetails struct associated with an operator.
   */
`,
	Example: `

pelldvs query operator operator-details --from <key-name> <operator-address>

pelldvs query operator operator-details --from pell-localnet-deployer 0xa0Ee7A142d267C1f36714E4a8F75612F20a79720
pelldvs query operator operator-details --from pell-localnet-deployer 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleQueryOperatorDetails(cmd, chainflags.FromKeyNameFlag.Value, args[0])
	},
}

func handleQueryOperatorDetails(cmd *cobra.Command, keyName string,
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

	result, err := execQueryOperatorDetails(cmd, kpath.ECDSA, operatorAddr)
	if err != nil {
		return fmt.Errorf("failed to create group: %v", err)
	}

	logger.Info("tx successfully", "result", result)

	return err
}

func execQueryOperatorDetails(cmd *cobra.Command, privKeyPath string, operator string) (*types.Operator, error) {
	logger.Info(
		"handleQueryOperatorDetails",
		"ethRPCURL", chainflags.EthRPCURLFlag.Value,
		"privKeyPath", privKeyPath,
		"operator", operator,
	)

	ctx := context.Background()
	address, err := ecdsa.GetAddressFromKeyStoreFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get address from key store file: %v", err)
	}

	chainOp, _, err := utils.NewOperatorFromFile(cmd, pellcfg.CmtConfig.Pell.InteractorConfigPath, logger)
	if err != nil {
		logger.Error("failed to create operator", "err", err)
		return nil, err
	}

	result, err := chainOp.GetOperatorDetails(&bind.CallOpts{Context: ctx}, operator)

	logger.Info(
		"handleQueryOperatorDetails",
		"address", address,
		"result", result,
	)

	return result, err
}

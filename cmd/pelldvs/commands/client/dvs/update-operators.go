package dvs

import (
	"context"
	"fmt"
	"strings"

	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

var updateOperatorsCmd = &cobra.Command{
	Use:   "update-operators",
	Short: "update-operators",
	Args:  cobra.ExactArgs(1),
	Long: `
  /**
   * @notice Updates the OperatorStakeManager's view of one or more operators' stakes. If any operator
   * is found to be below the minimum stake for the group, they are deregistered.
   * @dev stakes are queried from the PellNetwork core DelegationManager contract
   * @param operators a list of operator addresses to update
   */
`,
	Example: `
pelldvs client dvs update-operators --from <key-name> "<address1,address2,...>"
pelldvs client dvs update-operators --from pell-localnet-deployer "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266,0x14dC79964da2C08b23698B3D3cc7Ca32193d9955"

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		operators := args[0]
		return handleUpdateOperators(cmd, operators)
	},
}

func handleUpdateOperators(cmd *cobra.Command, operators string) error {
	operators = strings.TrimSpace(operators)
	operatorsList := strings.Split(operators, ",")

	// clean up the list, remove empty strings
	var cleanedOperatorsList []string
	for _, operator := range operatorsList {
		operator = strings.TrimSpace(operator)
		if operator != "" {
			cleanedOperatorsList = append(cleanedOperatorsList, operator)
		}
	}

	if len(cleanedOperatorsList) == 0 {
		return fmt.Errorf("no operators provided")
	}

	addresses := make([]string, 0, len(cleanedOperatorsList))
	for _, operator := range cleanedOperatorsList {
		operator = strings.TrimSpace(operator)
		if !gethcommon.IsHexAddress(operator) {
			return fmt.Errorf("invalid address %s", operator)
		}
		addresses = append(addresses, operator)
	}

	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	receipt, err := execUpdateOperators(cmd, kpath.ECDSA, addresses)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execUpdateOperators(cmd *cobra.Command, privKeyPath string, addresses []string) (*gethtypes.Receipt, error) {

	logger.Info(
		"execUpdateOperators",
		"ethRPCURL", chainflags.EthRPCURLFlag.Value,
		"privKeyPath", privKeyPath,
		"addresses", addresses,
	)

	ctx := context.Background()
	address, err := ecdsa.GetAddressFromKeyStoreFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get address from key store file: %v", err)
	}

	chainDVS, _, err := utils.NewDVSFromFromFile(cmd, pellcfg.CmtConfig.Pell.InteractorConfigPath, logger)
	if err != nil {
		logger.Error("failed to create operator", "err", err, "file", pellcfg.CmtConfig.Pell.InteractorConfigPath)
		return nil, err
	}

	receipt, err := chainDVS.UpdateOperators(ctx, addresses)

	logger.Info(
		"execUpdateOperators",
		"k", "v",
		"address", address,
		"receipt", receipt,
	)

	return receipt, err
}

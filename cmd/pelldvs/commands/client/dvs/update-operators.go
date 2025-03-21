package dvs

import (
	"context"
	"fmt"
	"strings"

	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs-libs/log"
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

pelldvs client dvs update-operators \
	--from <key-name> \
	--rpc-url <rpc-url> \
	--registry-router <registry-router> \
	"<address1,address2,...>"


`,
	Example: `
pelldvs client dvs update-operators \
	--from pell-localnet-deployer \
	--rpc-url http://localhost:8545 \
	--registry-router 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f \
	"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266,0x14dC79964da2C08b23698B3D3cc7Ca32193d9955"

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		operators := args[0]
		return handleUpdateOperators(cmd, operators)
	},
}

func handleUpdateOperators(cmd *cobra.Command, operators string) error {
	logger := getCmdLogger(cmd)
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

	receipt, err := execUpdateOperators(cmd, logger, kpath.ECDSA, addresses)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execUpdateOperators(cmd *cobra.Command, logger log.Logger, privKeyPath string, addresses []string) (*gethtypes.Receipt, error) {
	logger.Info(
		utils.GetPrettyCommandName(cmd),
		"privKeyPath", privKeyPath,
		"addresses", addresses,
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

	chainDVS, err := utils.NewDVSFromCfg(cfg, logger)
	if err != nil {
		logger.Error("failed to create chainDVS",
			"err", err,
			"file", pellcfg.CmtConfig.Pell.InteractorConfigPath,
			"cfg", fmt.Sprintf("%+v", cfg),
		)
		return nil, err
	}

	receipt, err := chainDVS.UpdateOperators(ctx, addresses)

	logger.Info(
		utils.GetPrettyCommandName(cmd),
		"k", "v",
		"address", address,
		"receipt", receipt,
	)

	return receipt, err
}

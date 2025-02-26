package operator

import (
	"context"
	"fmt"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
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
`,
	Example: `
pelldvs client operator deregister-operator --from <keyName>  <groupNumbers>
pelldvs client operator deregister-operator --from pell-localnet-deployer 1,2,3

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleDeRegisterOperator(cmd, chainflags.FromKeyNameFlag.Value, args[0])
	},
}

func handleDeRegisterOperator(cmd *cobra.Command, keyName string,
	groupNumbersStr string,
) error {

	groupNumbers := chainutils.ConvStrsToUint8List(groupNumbersStr)
	if len(groupNumbers) == 0 {
		return fmt.Errorf("invalid group numbers %s", groupNumbersStr)
	}
	kpath := keys.GetKeysPath(pellcfg.CmtConfig, keyName)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	receipt, err := execDeRegisterOperator(cmd, kpath.ECDSA, groupNumbers)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execDeRegisterOperator(cmd *cobra.Command, privKeyPath string, groupNumbers []byte) (*gethtypes.Receipt, error) {
	logger.Info(
		"handleDeRegisterOperator",
		"ethRPCURL", chainflags.EthRPCURLFlag.Value,
		"privKeyPath", privKeyPath,
		"operator", groupNumbers,
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

	receipt, err := chainOp.DeregisterOperator(ctx, groupNumbers)

	logger.Info(
		"handleDeRegisterOperator",
		"address", address,
		"receipt", receipt,
	)

	return receipt, err
}

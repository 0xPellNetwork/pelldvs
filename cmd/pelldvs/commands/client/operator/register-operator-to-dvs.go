package operator

import (
	"context"
	"fmt"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/bls"
	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainutils"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

var (
	registerOperatorToDVSFlagGroupNumbers = &chainflags.StringFlag{
		Name:  "groups",
		Usage: "group numbers",
	}
	registerOperatorToDVSFlagSocket = &chainflags.StringFlag{
		Name:  "socket",
		Usage: "socket",
	}
)

func init() {
	registerOperatorToDVSFlagGroupNumbers.AddToCmdFlag(registerOperatorToDVSCmd)
	registerOperatorToDVSFlagSocket.AddToCmdFlag(registerOperatorToDVSCmd)
	err := chainflags.MarkFlagsAreRequired(
		registerOperatorToDVSCmd,
		registerOperatorToDVSFlagGroupNumbers,
		registerOperatorToDVSFlagSocket,
	)
	if err != nil {
		panic(err)
	}
}

var registerOperatorToDVSCmd = &cobra.Command{
	Use:   "register-operator-to-dvs",
	Short: "register-operator-to-dvs",
	Long: `Registers msg.sender as an operator for one or more groups. If any group exceeds its maximum
   operator capacity after the operator is registered, this method will fail.

   * @param groupNumbers is an ordered byte array containing the group numbers being registered for
   * @param socket is the socket of the operator (typically an IP address)
`,
	Example: `

pelldvs client operator register-operator-to-dvs --from <key-name>  --groups <group-number> --socket <socket>

pelldvs client operator register-operator-to-dvs --from pell-localnet-deployer --groups 0 --socket http://127.0.0.1:8005

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleRegisterOperatorToDVS(
			cmd,
			registerOperatorToDVSFlagGroupNumbers.Value,
			registerOperatorToDVSFlagSocket.Value,
		)
	},
}

func handleRegisterOperatorToDVS(cmd *cobra.Command, groupNumbersStr string, socket string) error {
	groupNumbers := chainutils.ConvStrsToUint8List(groupNumbersStr)
	if len(groupNumbers) == 0 {
		return fmt.Errorf("invalid group numbers %s", groupNumbersStr)
	}

	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	receipt, err := execRegisterOperatorToDVS(cmd, kpath.ECDSA, kpath.BLS, groupNumbers, socket)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execRegisterOperatorToDVS(cmd *cobra.Command, privKeyPath, blsKeyPath string, groupNumbers []uint8, socket string) (*gethtypes.Receipt, error) {
	logger.Info(
		"handleRegisterOperatorToDVS",
		"ethRPCURL", chainflags.EthRPCURLFlag.Value,
		"privKeyPath", privKeyPath,
		"groupNumbers", groupNumbers,
		"socket", socket,
	)

	ctx := context.Background()
	address, err := ecdsa.GetAddressFromKeyStoreFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get address from key store file: %v", err)
	}

	blsKeyPair, err := bls.ReadPrivateKeyFromFile(blsKeyPath, "")
	if err != nil {
		return nil, fmt.Errorf("failed to read bls key: %v", err)
	}
	ecdsaPrivKey, err := ecdsa.ReadKey(privKeyPath, "")
	if err != nil {
		return nil, fmt.Errorf("failed to read ecdsa key: %v", err)
	}

	chainOp, _, err := utils.NewOperatorFromFile(cmd, pellcfg.CmtConfig.Pell.InteractorConfigPath, logger)
	if err != nil {
		logger.Error("failed to create operator", "err", err)
		return nil, err
	}

	receipt, err := chainOp.RegisterToDVS(ctx, ecdsaPrivKey, blsKeyPair, groupNumbers, socket)

	logger.Info(
		"handleRegisterOperatorToDVS",
		"address", address,
		"receipt", receipt,
	)

	return receipt, err
}

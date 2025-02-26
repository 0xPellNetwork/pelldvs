package operator

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

var (
	registerOperatorFlagMetadataURI = &chainflags.StringFlag{
		Name:  "metadata-uri",
		Usage: "metadata URI",
	}

	registerOperatorFlagApproverAddress = &chainflags.StringFlag{
		Name:  "approver",
		Usage: "approver address",
	}

	registerOperatorFlagStakerOptoutWindowSeconds = &chainflags.IntFlag{
		Name:  "staker-optout-window-seconds",
		Usage: "staker optout window seconds",
	}
)

func init() {
	registerOperatorFlagMetadataURI.AddToCmdFlag(registerOperatorCmd)
	registerOperatorFlagApproverAddress.AddToCmdFlag(registerOperatorCmd)
	registerOperatorFlagStakerOptoutWindowSeconds.AddToCmdFlag(registerOperatorCmd)

	err := chainflags.MarkFlagsAreRequired(registerOperatorCmd, registerOperatorFlagMetadataURI)
	if err != nil {
		panic(err)
	}
}

var registerOperatorCmd = &cobra.Command{
	Use:   "register-operator",
	Short: "register-operator",
	Long: `Registers msg.sender as an operator for one or more groups. If any group exceeds its maximum
   operator capacity after the operator is registered, this method will fail.

   * @param groupNumbers is an ordered byte array containing the group numbers being registered for
   * @param socket is the socket of the operator (typically an IP address)
`,
	Example: `

pelldvs client operator register-operator --from <key-name>  --metadata-uri <metadata-uri>

pelldvs client operator register-operator --from pell-localnet-deployer --metadata-uri https://example.com/metadata.json

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleRegisterOperator(
			cmd,
			chainflags.FromKeyNameFlag.Value,
			registerOperatorFlagMetadataURI.Value,
			uint32(registerOperatorFlagStakerOptoutWindowSeconds.Value),
			registerOperatorFlagApproverAddress.Value,
		)
	},
}

func handleRegisterOperator(cmd *cobra.Command, keyName string, metadataURI string, stakerOptoutWindowSeconds uint32, approverAddress string) error {

	kpath := keys.GetKeysPath(pellcfg.CmtConfig, keyName)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	receipt, err := execRegisterOperator(cmd, kpath.ECDSA, metadataURI, stakerOptoutWindowSeconds, approverAddress)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execRegisterOperator(cmd *cobra.Command, privKeyPath string, metadataURI string, stakerOptoutWindowSeconds uint32, approverAddress string) (*gethtypes.Receipt, error) {
	logger.Info(
		"handleRegisterOperator",
		"ethRPCURL", chainflags.EthRPCURLFlag.Value,
		"privKeyPath", privKeyPath,
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

	isRegistered, err := chainOp.IsOperator(&bind.CallOpts{Context: ctx}, address.String())
	if err != nil {
		return nil, fmt.Errorf("failed to check if operator already registered: %v", err)
	}
	if isRegistered {
		return nil, fmt.Errorf("[stop] operator %s already registered", address.String())
	}

	receipt, err := chainOp.RegisterAsOperator(ctx, address.String(), stakerOptoutWindowSeconds, approverAddress, metadataURI)

	logger.Info(
		"handleRegisterOperator",
		"address", address,
		"receipt", receipt,
	)

	return receipt, err
}

package operator

import (
	"context"
	"fmt"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

var updateSocketCmd = &cobra.Command{
	Use:   "update-socket",
	Short: "update-socket",
	Args:  cobra.ExactArgs(1),
	Long: `Updates the socket of the msg.sender given they are a registered operator
   * @param socket is the new socket of the operator`,
	Example: `

pelldvs client operator update-socket --from <key-name>  <socket-uri>
pelldvs client operator update-socket --from pell-localnet-deployer '127.0.0.1:9988'
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		uri := args[0]
		return handleUpdateSocket(cmd, uri)
	},
}

func handleUpdateSocket(cmd *cobra.Command, socket string) error {
	keyName := chainflags.FromKeyNameFlag.GetValue()

	kpath := keys.GetKeysPath(pellcfg.CmtConfig, keyName)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	receipt, err := execUpdateSocket(cmd, kpath.ECDSA, socket)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execUpdateSocket(cmd *cobra.Command, privKeyPath string, uri string) (*gethtypes.Receipt, error) {
	logger.Info(
		"handleUpdateSocket",
		"ethRPCURL", chainflags.EthRPCURLFlag.Value,
		"privKeyPath", privKeyPath,
		"uri", uri,
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

	receipt, err := chainOp.UpdateSocket(ctx, uri)

	logger.Info(
		"handleUpdateSocket",
		"address", address,
		"receipt", receipt,
	)

	return receipt, err
}

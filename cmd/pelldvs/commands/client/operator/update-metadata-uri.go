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

var updateOperatorMetadataURICmd = &cobra.Command{
	Use:   "update-metadata-uri",
	Short: "update-metadata-uri",
	Args:  cobra.ExactArgs(1),
	Long: ` Called by an operator to emit an OperatorMetadataURIUpdated event indicating the information has updated.
   * @param metadataURI The URI for metadata associated with an operator
`,

	Example: `

pelldvs client operator update-metadata-uri --from <key-name> <metadataURI>
pelldvs client operator update-metadata-uri --from pell-localnet-deployer 'https://raw.githubusercontent.com/a/b/c.json'

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		uri := args[0]
		return handleUpdateOperatorMetadataURI(cmd, chainflags.FromKeyNameFlag.Value, uri)
	},
}

func handleUpdateOperatorMetadataURI(cmd *cobra.Command, keyName string, uri string) error {
	kpath := keys.GetKeysPath(pellcfg.CmtConfig, keyName)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	receipt, err := execUpdateOperatorMetadataURI(cmd, kpath.ECDSA, uri)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execUpdateOperatorMetadataURI(cmd *cobra.Command, privKeyPath string, uri string) (*gethtypes.Receipt, error) {
	logger.Info(
		"handleUpdateOperatorMetadataURI",
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

	receipt, err := chainOp.UpdateMetadataURI(ctx, uri)

	logger.Info(
		"handleUpdateOperatorMetadataURI",
		"address", address,
		"receipt", receipt,
	)

	return receipt, err
}

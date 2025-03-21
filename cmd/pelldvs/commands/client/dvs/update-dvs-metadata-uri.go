package dvs

import (
	"context"
	"fmt"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

var updateDVSMetadataURICmd = &cobra.Command{
	Use:   "update-dvs-metadata-uri",
	Short: "update-dvs-metadata-uri",
	Args:  cobra.ExactArgs(1),
	Long: `
  /**
   * @notice Updates the metadata URI for the DVS
   * @param _metadataURI is the metadata URI for the DVS
   * Metadata should follow the format outlined by this example.
        {
            "name": "PellNetwork DVS 1",
            "website": "https://pell.network/",
            "description": "This is my 1st DVS",
            "logo": "https://pell.network/static/media/service-bg.6cdb83513c3e4ce7d288.png",
            "twitter": "https://twitter.com/Pell_Network"
        }
   * @dev only callable by the owner
   */

pelldvs client dvs update-dvs-metadata-uri \
	--from <key-name> \
	--rpc-url <rpc-url> \
	--registry-router <registry-router> \
	<uri>

`,
	Example: `
pelldvs client dvs update-dvs-metadata-uri \
	--from pell-localnet-deployer \
	--rpc-url http://localhost:8545 \
	--registry-router 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f \
	https://raw.githubusercontent.com/example/repo/file.json

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		uri := args[0]
		return handleUpdateDVSMetadataURI(cmd, chainflags.FromKeyNameFlag.Value, uri)
	},
}

func handleUpdateDVSMetadataURI(cmd *cobra.Command, keyName string, uri string) error {
	logger := getCmdLogger(cmd)
	kpath := keys.GetKeysPath(pellcfg.CmtConfig, keyName)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	receipt, err := execUpdateDVSMetadataURI(cmd, logger, kpath.ECDSA, uri)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execUpdateDVSMetadataURI(cmd *cobra.Command, logger log.Logger, privKeyPath string, uri string) (*gethtypes.Receipt, error) {
	logger.Info(
		utils.GetPrettyCommandName(cmd),
		"privKeyPath", privKeyPath,
		"uri", uri,
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

	receipt, err := chainDVS.UpdateDVSMetadataURI(ctx, uri)

	logger.Info(
		utils.GetPrettyCommandName(cmd),
		"sender", address,
		"receipt", receipt,
	)

	return receipt, err
}

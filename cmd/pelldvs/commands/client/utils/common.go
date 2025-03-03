package utils

import (
	"fmt"

	"github.com/spf13/cobra"

	chaincfg "github.com/0xPellNetwork/pelldvs-interactor/config"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs-libs/os"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
)

func LoadChainConfig(cmd *cobra.Command, filePath string, logger log.Logger) (*chaincfg.Config, error) {
	if !os.FileExists(filePath) {
		logger.Info("Chain config file not found, use default chain config")
		cfg := chaincfg.DefaultConfig()
		UpdateChainConfigFromFlags(cmd, cfg, logger)
		return cfg, nil
	}
	cfg, err := chaincfg.LoadConfig(filePath)
	if err != nil {
		logger.Error("Failed to load chain config from file, use default chain config", "error", err)
		return nil, err
	}

	UpdateChainConfigFromFlags(cmd, cfg, logger)

	return cfg, nil
}

func UpdateChainConfigFromFlags(cmd *cobra.Command, cfg *chaincfg.Config, logger log.Logger) {
	logger.Debug("chaincfg before overwrite", "cfg", fmt.Sprintf("%+v", cfg))

	if cmd.Flags().Lookup(chainflags.EthRPCURLFlag.Name).Changed {
		cfg.RPCURL = chainflags.EthRPCURLFlag.GetValue()
	}
	if cmd.Flags().Lookup(chainflags.PellRegistryRouterFactoryAddress.Name).Changed {
		cfg.ContractConfig.PellRegistryRouterFactory = chainflags.PellRegistryRouterFactoryAddress.GetValue()
	}

	if cmd.Flags().Lookup(chainflags.PellRegistryRouterAddress.Name).Changed {
		cfg.ContractConfig.PellRegistryRouter = chainflags.PellRegistryRouterAddress.GetValue()
	}

	if cmd.Flags().Lookup(chainflags.PellDelegationManagerAddress.Name).Changed {
		cfg.ContractConfig.PellDelegationManager = chainflags.PellDelegationManagerAddress.GetValue()
	}
	if cmd.Flags().Lookup(chainflags.PellDVSDirectoryAddress.Name).Changed {
		cfg.ContractConfig.PellDVSDirectory = chainflags.PellDVSDirectoryAddress.GetValue()
	}

	keyName := cmd.Flags().Lookup(chainflags.FromKeyNameFlag.Name).Value.String()
	if keyName != "" {
		cfg.ECDSAPrivateKeyFilePath = fmt.Sprintf("%s/keys/%s.ecdsa.key.json", pellcfg.CmtConfig.RootDir, keyName)
	}

	logger.Debug("chaincfg after overwrite", "cfg", fmt.Sprintf("%+v", cfg))
}

func GetPrettyCommandName(cmd *cobra.Command) string {
	if cmd.HasParent() {
		return fmt.Sprintf("%s/%s", cmd.Parent().Use, cmd.Use)
	}
	return cmd.Use
}

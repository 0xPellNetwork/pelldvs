package utils

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	chaincfg "github.com/0xPellNetwork/pelldvs-interactor/config"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
)

func LoadChainConfig(cmd *cobra.Command, filePath string, logger log.Logger) (*chaincfg.Config, error) {
	cfg, err := chaincfg.LoadConfig(filePath)
	if err != nil {
		logger.Info("Failed to load chain config from file, use default chain config", "error", err)
		cfg = &chaincfg.Config{
			RPCURL: "http://localhost:8545",
			ContractConfig: chaincfg.ContractConfig{
				DVSConfigs:            make(map[uint64]*chaincfg.DVSConfig),
				IndexerBatchSize:      100,
				IndexerListenInterval: 5 * time.Second,
			},
		}
	}

	UpdateChainConfigFromFlags(cmd, cfg, logger)

	return cfg, nil
}

func UpdateChainConfigFromFlags(cmd *cobra.Command, cfg *chaincfg.Config, logger log.Logger) {
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

	keyName := cmd.Flags().Lookup(chainflags.FromKeyNameFlag.Name).Value
	cfg.ECDSAPrivateKeyFilePath = fmt.Sprintf("%s/keys/%s.ecdsa.key.json", pellcfg.CmtConfig.RootDir, keyName)

	logger.Debug("chaincfg after overwrite", "cfg", fmt.Sprintf("%+v", cfg))
}

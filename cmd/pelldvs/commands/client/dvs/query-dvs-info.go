package dvs

import (
	"context"
	"fmt"

	gethbind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/spf13/cobra"

	chaintypes "github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

func init() {
	chainflags.ChainIDFlag.AddToCmdFlag(queryDVSInfoCmd)

	err := chainflags.MarkFlagsAreRequired(queryDVSInfoCmd, chainflags.ChainIDFlag)
	if err != nil {
		panic(err)
	}
}

var queryDVSInfoCmd = &cobra.Command{
	Use:   "query-dvs-info",
	Short: "query-dvs-info",
	Long: `
pelldvs client dvs query-dvs-info \
	--from <key-name> \
	--rpc-url <rpc-url> \
	--registry-router <registry-router> \
	--chain_id <chain-id>
`,
	Example: `
pelldvs client dvs query-dvs-info \
	--from pell-localnet-deployer \
	--rpc-url http://localhost:8545 \
	--registry-router 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f \
	--chain_id 666
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleQueryDVSInfo(cmd)
	},
}

func handleQueryDVSInfo(cmd *cobra.Command) error {
	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	dvsInfo, err := execQueryDVSInfo(cmd, chainflags.ChainIDFlag.Value)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "result", fmt.Sprintf("%+v", dvsInfo))

	return err
}

func execQueryDVSInfo(cmd *cobra.Command, chainID int) (*chaintypes.DVSInfo, error) {
	cmdName := utils.GetPrettyCommandName(cmd)
	logger.Info(cmdName,
		"chainID", chainID,
	)

	ctx := context.Background()
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

	dvsInfo, err := chainDVS.QueryDVSInfo(&gethbind.CallOpts{Context: ctx}, uint64(chainID))

	logger.Info(cmdName,
		"k", "v",
		"res", dvsInfo,
	)

	return dvsInfo, err
}

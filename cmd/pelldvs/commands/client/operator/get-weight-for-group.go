package operator

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-interactor/chainlibs/eth"
	chaincfg "github.com/0xPellNetwork/pelldvs-interactor/config"
	chaintypes "github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainutils"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
)

func init() {
	chainflags.DVSRPCURL.AddToCmdFlag(getWeightForGroupCmd)
	chainflags.DVSOperatorStakeManagerAddress.AddToCmdFlag(getWeightForGroupCmd)

	err := chainflags.MarkFlagsAreRequired(getWeightForGroupCmd,
		chainflags.DVSRPCURL,
		chainflags.DVSOperatorStakeManagerAddress,
	)
	if err != nil {
		panic(err)
	}
}

var getWeightForGroupCmd = &cobra.Command{
	Use:   "get-weight-for-group",
	Short: "get-weight-for-group",
	Args:  cobra.ExactArgs(2),
	Long: `
  /**
   * @notice Returns true is an operator has previously registered for delegation.
   */

pelldvs client operator get-weight-for-group \
	--rpc-url <rpc-url, optional> \
	--registry-router <registry-router, optional > \
	--dvs-operator-stake-manager <operator-stake-manager> \
	--dvs-rpc-url <dvs-rpc-url> \
	<group-number> <operator-address>

`,
	Example: `
pelldvs client operator get-weight-for-group \
	--rpc-url http://localhost:8646 \
	--dvs-rpc-url http://localhost:8646 \
	--registry-router 0xE7402c51ae0bd667ad57a99991af6C2b686cd4f1 \
	--dvs-operator-stake-manager 0x8198f5d8F8CfFE8f9C413d98a0A55aEB8ab9FbB7 \
	0 0xac4f337423254816ee81cd3b23335fc1bc8b36f1

pelldvs client operator get-weight-for-group \
	--dvs-rpc-url http://localhost:8646 \
	--dvs-operator-stake-manager 0x8198f5d8F8CfFE8f9C413d98a0A55aEB8ab9FbB7 \
	0 0xac4f337423254816ee81cd3b23335fc1bc8b36f1

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleGetWeightForGroup(cmd, args[0], args[1])
	},
}

func handleGetWeightForGroup(cmd *cobra.Command, groupNumberStr, operatorAddr string) error {
	logger := getCmdLogger(cmd)

	if !gethcommon.IsHexAddress(operatorAddr) {
		return fmt.Errorf("invalid address %s", operatorAddr)
	}

	groupNumber, err := chainutils.ConvStrToUint8(groupNumberStr)
	if err != nil {
		return fmt.Errorf("invalid group number %s", groupNumberStr)
	}

	result, err := execGetWeightForGroup(cmd, logger, groupNumber, operatorAddr)
	if err != nil {
		return fmt.Errorf("failed to execGetWeightForGroup: %v", err)
	}

	logger.Info("tx successfully", "result", fmt.Sprintf("%+v", result))

	return err
}

func execGetWeightForGroup(cmd *cobra.Command, logger log.Logger, groupNumber uint8, operatorAddr string) (*chaintypes.OperatorWeightForGroup, error) {
	logger.Info(
		"exec start",
		"groupNumber", groupNumber,
		"operatorAddr", operatorAddr,
	)

	cfg, err := utils.LoadChainConfig(cmd, pellcfg.CmtConfig.Pell.InteractorConfigPath, logger)
	if err != nil {
		logger.Error("failed to load chain config",
			"err", err, "file", pellcfg.CmtConfig.Pell.InteractorConfigPath,
		)
		return nil, err
	}

	logger.Info("chain config details", "chaincfg", fmt.Sprintf("%+v", cfg))

	var chainConfigChecker = utils.NewChainConfigChecker(cfg)
	if !chainConfigChecker.HasRPCURL() {
		logger.Info("chain rpc url not provided")
	}
	if !chainConfigChecker.IsValidPellRegistryRouter() {
		logger.Info("pell registry router not provided")
	}
	if chainflags.DVSRPCURL.Value == "" {
		return nil, fmt.Errorf("dvs rpc url is required")
	}

	dvsETHClient, err := eth.NewClient(chainflags.DVSRPCURL.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to create eth client for RPCURL: %s:, err %v",
			chainflags.DVSRPCURL.Value, err,
		)
	}

	dvsChainID, err := dvsETHClient.ChainID(cmd.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain id from rpc client: url:%s , error: %v",
			chainflags.DVSRPCURL.Value, err,
		)
	}

	dvsConfig := chaincfg.DVSConfig{
		ChainID:              dvsChainID.Uint64(),
		RPCURL:               chainflags.DVSRPCURL.Value,
		OperatorStakeManager: chainflags.DVSOperatorStakeManagerAddress.Value,
	}

	cfg.ContractConfig.DVSConfigs[dvsChainID.Uint64()] = &dvsConfig
	logger.Info("dvs config", "dvsConfig", dvsConfig)

	chainOp, err := utils.NewOperatorFromCfg(cfg, logger)
	if err != nil {
		logger.Error("failed to create chain operator",
			"err", err,
			"file", pellcfg.CmtConfig.Pell.InteractorConfigPath,
			"cfg", fmt.Sprintf("%+v", cfg),
			"DVSConfigs", fmt.Sprintf("%+v", cfg.ContractConfig.DVSConfigs[dvsChainID.Uint64()]),
		)
		return nil, err
	}

	ctx := context.Background()
	result, err := chainOp.GetWeightOfOperatorForGroup(
		&bind.CallOpts{Context: ctx},
		dvsChainID.Uint64(), // chain id
		groupNumber, gethcommon.HexToAddress(operatorAddr),
	)

	logger.Info(
		"exec done",
		"address", operatorAddr,
		"groupNumber", groupNumber,
		"result", result,
	)

	return result, err
}

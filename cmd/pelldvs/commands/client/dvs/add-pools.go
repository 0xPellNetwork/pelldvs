package dvs

import (
	"context"
	"fmt"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	chaintypes "github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	cmtos "github.com/0xPellNetwork/pelldvs-libs/os"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainutils"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

var addPoolsCmdFlagGroupNumber = &chainflags.StringFlag{
	Name:  "group",
	Usage: "group number",
}

var addPoolsFlagConfigFile = &chainflags.StringFlag{
	Name:  "config",
	Usage: "config file path",
}

func init() {
	addPoolsFlagConfigFile.AddToCmdFlag(addPoolsCmd)
	addPoolsCmdFlagGroupNumber.AddToCmdFlag(addPoolsCmd)

	err := chainflags.MarkFlagsAreRequired(addPoolsCmd, addPoolsCmdFlagGroupNumber, addPoolsFlagConfigFile)
	if err != nil {
		panic(err)
	}
}

var addPoolsCmd = &cobra.Command{
	Use:   "add-pools",
	Short: "add pools",
	Long: `
pelldvs client dvs add-pools \
	--from <key-name> \
	--rpc-url <rpc-url> \
	--registry-router <registry-router> \
	--group <group> \
	--config <group-config-json-path>
`,
	Example: `
pelldvs client dvs add-pools \
	--from pell-localnet-deployer \
	--rpc-url http://localhost:8545 \
	--registry-router 0x1234567890123456789012345678901234567890 \
	--group 1 \
	--config /path/to/group-config.json
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleAddPools(cmd)
	},
}

func handleAddPools(cmd *cobra.Command) error {
	logger := getCmdLogger(cmd)
	groupNumber, err := chainutils.ConvStrToUint8(addPoolsCmdFlagGroupNumber.Value)
	if err != nil {
		return fmt.Errorf("failed to convert group number %s to uint8: %v", addPoolsCmdFlagGroupNumber.Value, err)
	}

	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	// TODO(jimmy):  password support
	if !cmtos.FileExists(addPoolsFlagConfigFile.Value) {
		return fmt.Errorf("param file does not exist %s", addPoolsFlagConfigFile.Value)
	}

	var addPoolsParam chaintypes.AddPoolsRequest
	err = decodeJSONFromFile(addPoolsFlagConfigFile.Value, &addPoolsParam)
	if err != nil {
		return fmt.Errorf("failed to unmarshal addPoolsParam: %v", err)
	}

	hasSetGroupNumber := cmd.Flags().Lookup(addPoolsCmdFlagGroupNumber.Name).Changed
	if hasSetGroupNumber {
		addPoolsParam.GroupNumber = groupNumber
	}

	receipt, err := execAddPools(cmd, logger, addPoolsParam, kpath.ECDSA)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execAddPools(cmd *cobra.Command, logger log.Logger, params chaintypes.AddPoolsRequest, privKeyPath string) (*gethtypes.Receipt, error) {
	cmdName := utils.GetPrettyCommandName(cmd)
	logger.Info(fmt.Sprintf("%s start", cmdName),
		"privKeyPath", privKeyPath,
		"groupNumber", params.GroupNumber,
		"params", params,
	)

	ctx := context.Background()

	senderAddress, err := ecdsa.GetAddressFromKeyStoreFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get senderAddress from key store file: %v", err)
	}
	logger.Info(cmdName,
		"sender", senderAddress,
	)

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

	receipt, err := chainDVS.AddPools(ctx, params.GroupNumber, &params)
	if err != nil {
		return nil, err
	}

	logger.Info(
		fmt.Sprintf("%s done", cmdName),
		"k", "v",
		"senderAddress", senderAddress,
		"receipt", receipt,
	)

	return receipt, err
}

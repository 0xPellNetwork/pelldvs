package dvs

import (
	"context"
	"fmt"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	chaintypes "github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	cmtos "github.com/0xPellNetwork/pelldvs/libs/os"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

var groupConfigFileFlag = chainflags.StringFlag{
	Name:  "config",
	Usage: "group config file path",
}

func init() {
	createGroupCmd.Flags().StringVar(&groupConfigFileFlag.Value, groupConfigFileFlag.Name, "", groupConfigFileFlag.Usage)
	err := chainflags.MarkFlagsAreRequired(createGroupCmd, &groupConfigFileFlag)
	if err != nil {
		panic(err)
	}
}

var createGroupCmd = &cobra.Command{
	Use:   "create-group",
	Short: "Create a group",
	Long: `Create a group
   * @notice Creates a group and initializes it in each registry contract
   * @param operatorSetParams configures the group's max operator count and churn parameters
   * @param minimumStake sets the minimum stake required for an operator to register or remain
   * registered
`,
	Example: `
pelldvs client dvs create-group --from <key-name> --config <config-json-path>
pelldvs client dvs create-group --from pell-localnet-deployer --config /data/pells/dvsreqs2/create-group-1.json
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleCreateGroup(cmd, chainflags.FromKeyNameFlag.Value, groupConfigFileFlag.Value)
	},
}

func handleCreateGroup(cmd *cobra.Command, keyName, paramFilePath string) error {
	kpath := keys.GetKeysPath(pellcfg.CmtConfig, keyName)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	if !cmtos.FileExists(paramFilePath) {
		return fmt.Errorf("param file does not exist %s", paramFilePath)
	}

	var createGroupParam chaintypes.CreateGroupRequest
	err := decodeJSONFromFile(paramFilePath, &createGroupParam)
	if err != nil {
		return fmt.Errorf("failed to unmarshal createGroupParam: %v", err)
	}

	receipt, err := execCreateGroup(cmd, createGroupParam, kpath.ECDSA)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execCreateGroup(cmd *cobra.Command, params chaintypes.CreateGroupRequest, privKeyPath string) (*gethtypes.Receipt, error) {
	cmdName := "createGroupParam"

	logger.Info(fmt.Sprintf("%s start", cmdName),
		"privKeyPath", privKeyPath,
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

	chainDVS, _, err := utils.NewDVSFromFromFile(cmd, pellcfg.CmtConfig.Pell.InteractorConfigPath, logger)
	if err != nil {
		logger.Error("failed to create operator", "err", err, "file", pellcfg.CmtConfig.Pell.InteractorConfigPath)
		return nil, err
	}

	groupCount, err := chainDVS.QueryGroupCount(ctx)
	if err != nil {
		logger.Error("failed to query group count", "err", err)
	}

	logger.Info("group count before", "count", groupCount)

	receipt, err := chainDVS.CreateGroup(ctx, &params)
	if err != nil {
		return nil, err
	}

	logger.Info(
		fmt.Sprintf("%s done", cmdName),
		"k", "v",
		"senderAddress", senderAddress,
		"receipt", receipt,
	)

	groupCount, err = chainDVS.QueryGroupCount(ctx)
	if err != nil {
		logger.Error("failed to query group count", "err", err)
	}

	logger.Info("group count after", "count", groupCount)

	return receipt, err
}

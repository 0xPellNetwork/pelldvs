package dvs

import (
	"context"
	"fmt"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	chaintypes "github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainutils"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	cmtos "github.com/0xPellNetwork/pelldvs/libs/os"
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
	Example: `
pelldvs client dvs add-pools --from <key-name> --group 0 --config <group-config-json-path>
pelldvs client dvs add-pools --from pell-localnet-deployer --group 0 --config /data/pells/dvsreqs2/create-group-1.json
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		groupNumber, err := chainutils.ConvStrToUint8(addPoolsCmdFlagGroupNumber.Value)
		if err != nil {
			return fmt.Errorf("failed to convert group number %s to uint8: %v", addPoolsCmdFlagGroupNumber.Value, err)
		}
		return handleAddPools(cmd,
			groupNumber,
			addPoolsFlagConfigFile.Value,
		)
	},
}

func handleAddPools(cmd *cobra.Command, groupNumber uint8, paramFilePath string) error {
	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	// TODO(jimmy):  password support
	if !cmtos.FileExists(paramFilePath) {
		return fmt.Errorf("param file does not exist %s", paramFilePath)
	}

	var addPoolsParam chaintypes.AddPoolsRequest
	err := decodeJSONFromFile(paramFilePath, &addPoolsParam)
	if err != nil {
		return fmt.Errorf("failed to unmarshal addPoolsParam: %v", err)
	}

	hasSetGroupNumber := cmd.Flags().Lookup(addPoolsCmdFlagGroupNumber.Name).Changed
	if hasSetGroupNumber {
		addPoolsParam.GroupNumber = groupNumber
	}

	receipt, err := execAddPools(cmd, addPoolsParam, kpath.ECDSA)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execAddPools(cmd *cobra.Command, params chaintypes.AddPoolsRequest, privKeyPath string) (*gethtypes.Receipt, error) {
	cmdName := "handleAddPools"

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

	chainDVS, _, err := utils.NewDVSFromFromFile(cmd, pellcfg.CmtConfig.Pell.InteractorConfigPath, logger)
	if err != nil {
		logger.Error("failed to create operator", "err", err, "file", pellcfg.CmtConfig.Pell.InteractorConfigPath)
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

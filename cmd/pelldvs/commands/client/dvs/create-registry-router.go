package dvs

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/spf13/cobra"

	chaintypes "github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	cmtos "github.com/0xPellNetwork/pelldvs/libs/os"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

var createRegistryRouterFlagInitOwner = &chainflags.StringFlag{
	Name:  "initial-owner",
	Usage: "the initial owner of the contract",
}

var createRegistryRouterFlagDVSChainApprover = &chainflags.StringFlag{
	Name:  "dvs-chain-approver",
	Usage: "the churn approver of the contract",
}

var createRegistryRouterFlagChurnApprover = &chainflags.StringFlag{
	Name:  "churn-approver",
	Usage: "the churn approver of the contract",
}

var createRegistryRouterFlagEjector = &chainflags.StringFlag{
	Name:  "ejector",
	Usage: "the ejector of the contract",
}

var createRegistryRouterFlagPauser = &chainflags.StringFlag{
	Name:  "pauser",
	Usage: "the pauser of the contract",
}

var createRegistryRouterFlagUnpauser = &chainflags.StringFlag{
	Name:  "unpauser",
	Usage: "the unpauser of the contract",
}

var createRegistryRouterFlagInitialPausedStatus = &chainflags.StringFlag{
	Name:  "initial-paused-status",
	Usage: "the initial paused status of the contract",
}

// save to file flag
var createRegistryRouterFlagSaveToFile = &chainflags.StringFlag{
	Name:  "save-to-file",
	Usage: "save the contract address to file",
}

// save to file flag
var createRegistryRouterFlagForceSave = &chainflags.StringFlag{
	Name:  "force-save",
	Usage: "force save the contract address to file",
}

func init() {
	createRegistryRouterCmd.Flags().StringVar(&createRegistryRouterFlagInitOwner.Value, createRegistryRouterFlagInitOwner.Name, "", createRegistryRouterFlagInitOwner.Usage)
	createRegistryRouterCmd.Flags().StringVar(&createRegistryRouterFlagDVSChainApprover.Value, createRegistryRouterFlagDVSChainApprover.Name, "", createRegistryRouterFlagDVSChainApprover.Usage)
	createRegistryRouterCmd.Flags().StringVar(&createRegistryRouterFlagChurnApprover.Value, createRegistryRouterFlagChurnApprover.Name, "", createRegistryRouterFlagChurnApprover.Usage)
	createRegistryRouterCmd.Flags().StringVar(&createRegistryRouterFlagEjector.Value, createRegistryRouterFlagEjector.Name, "", createRegistryRouterFlagEjector.Usage)
	createRegistryRouterCmd.Flags().StringVar(&createRegistryRouterFlagPauser.Value, createRegistryRouterFlagPauser.Name, "", createRegistryRouterFlagPauser.Usage)
	createRegistryRouterCmd.Flags().StringVar(&createRegistryRouterFlagUnpauser.Value, createRegistryRouterFlagUnpauser.Name, "", createRegistryRouterFlagUnpauser.Usage)
	createRegistryRouterCmd.Flags().StringVar(&createRegistryRouterFlagInitialPausedStatus.Value, createRegistryRouterFlagInitialPausedStatus.Name, "", createRegistryRouterFlagInitialPausedStatus.Usage)
	createRegistryRouterCmd.Flags().StringVar(&createRegistryRouterFlagSaveToFile.Value, createRegistryRouterFlagSaveToFile.Name, "", createRegistryRouterFlagSaveToFile.Usage)
	createRegistryRouterCmd.Flags().StringVar(&createRegistryRouterFlagForceSave.Value, createRegistryRouterFlagForceSave.Name, "", createRegistryRouterFlagForceSave.Usage)

	err := chainflags.MarkFlagsAreRequired(createRegistryRouterCmd, createRegistryRouterFlagInitOwner, createRegistryRouterFlagChurnApprover, createRegistryRouterFlagEjector, createRegistryRouterFlagPauser, createRegistryRouterFlagUnpauser)
	if err != nil {
		panic(err)
	}
}

var createRegistryRouterCmd = &cobra.Command{
	Use:   "create-registry-router",
	Short: "Create RegistryRouter",
	Long: `
pelldvs client dvs create-registry-router \
	--from <owner_key> \
    --rpc-url <eth rpc url> \
	--registry-router-factory <registry_router_factory_address> \
    --initial-owner <owner_address> \
    --dvs-chain-approver <dvs_chain_approver_address> \
    --churn-approver <churn_approver_address> \
    --ejector <ejector_address> \
    --pauser <pauser_address> \
    --unpauser <unpauser_address> \
    --initial-paused-status false \
	--save-to-file /path/to/registryRouterAddress.json \
	--force-save true
`,
	Example: `
pelldvs client dvs create-registry-router \
	--from pell-localnet-deployer \
	--rpc-url http://localhost:8545 \
	--registry-router-factory 0x1234567890123456789012345678901234567890 \
	--initial-owner 0x1234567890123456789012345678901234567890 \
	--dvs-chain-approver 0x1234567890123456789012345678901234567890 \
	--churn-approver 0x1234567890123456789012345678901234567890 \
	--ejector 0x1234567890123456789012345678901234567890 \
	--pauser 0x1234567890123456789012345678901234567890 \
	--unpauser 0x1234567890123456789012345678901234567890 \
	--initial-paused-status false \
	--save-to-file /tmp/registryRouterAddress.json \
	--force-save true
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleCreateRegistryRouter(cmd)
	},
}

func parseBoolValueFromString(s string) bool {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	for _, v := range []string{"true", "t", "yes", "y", "1"} {
		if s == v {
			return true
		}
	}

	return false
}

func handleCreateRegistryRouter(cmd *cobra.Command) error {
	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	// check if file is exist
	if createRegistryRouterFlagSaveToFile.Value != "" {
		if cmtos.FileExists(createRegistryRouterFlagSaveToFile.Value) {
			if !parseBoolValueFromString(createRegistryRouterFlagForceSave.Value) {
				return fmt.Errorf("file already exists %s", createRegistryRouterFlagSaveToFile.Value)
			}
		}
	}

	createRegistryRouterParam := &chaintypes.CreateRegistryRouterRequest{
		InitialOwner:        createRegistryRouterFlagInitOwner.Value,
		DVSChainApprover:    createRegistryRouterFlagDVSChainApprover.Value,
		ChurnApprover:       createRegistryRouterFlagChurnApprover.Value,
		Ejector:             createRegistryRouterFlagEjector.Value,
		Pausers:             []string{createRegistryRouterFlagPauser.Value},
		Unpauser:            createRegistryRouterFlagUnpauser.Value,
		InitialPausedStatus: big.NewInt(0),
	}

	if parseBoolValueFromString(createRegistryRouterFlagInitialPausedStatus.Value) {
		createRegistryRouterParam.InitialPausedStatus = big.NewInt(1)
	}

	res, err := execCreateRegistryRouter(cmd, createRegistryRouterParam, kpath.ECDSA)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", res.Receipt.TxHash.String())
	logger.Info("tx successfully", "registryRouterAddress: ", res.Address)

	// if filepath is empty, save to default file
	if createRegistryRouterFlagSaveToFile.Value == "" {
		createRegistryRouterFlagSaveToFile.Value = fmt.Sprintf("/tmp/pell-registryRouterAddress-%d-%s.json", time.Now().UnixNano(), res.Address)
	}

	bdata, err := json.Marshal(map[string]string{"address": res.Address})
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}
	err = cmtos.WriteFile(createRegistryRouterFlagSaveToFile.Value, bdata, 0644)
	if err != nil {
		return fmt.Errorf("failed to save to file: %v", err)
	}
	logger.Info("save to file successfully", "file: ", createRegistryRouterFlagSaveToFile.Value)

	return err
}

func execCreateRegistryRouter(cmd *cobra.Command,
	params *chaintypes.CreateRegistryRouterRequest,
	privKeyPath string,
) (*chaintypes.CreateRegistryRouterResponse, error) {
	cmdName := utils.GetPrettyCommandName(cmd)
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
	if !chainConfigChecker.IsValidPellRegistryRouterFactory() {
		return nil, fmt.Errorf("pell registry router factory is required")
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

	resp, err := chainDVS.CreateRegistryRouter(ctx, params)
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("%s done", cmdName),
		"k", "v",
		"senderAddress", senderAddress,
		"resp", resp,
	)

	return resp, err
}

package operator

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainutils"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

var modifyOperatorDetailsCmd = &cobra.Command{
	Use:   "modify-operator-details",
	Short: "modify-operator-details",
	Args:  cobra.MinimumNArgs(1),
	Long: ` Modify operator details.

staker-optout-window-seconds[required]： A minimum delay -- enforced between:
  1) the operator signaling their intent to register for a service, via calling Slasher.optIntoSlashing
  2) the operator completing registration for the service, via the service ultimately calling "Slasher.recordFirstStakeUpdate"
  note that for a specific operator, this value *cannot decrease*, i.e. if the operator wishes to modify their OperatorDetails,
  then they are only allowed to either increase this value or keep it the same.

metadataURI[optional]: 
  is a URI for the operator's metadata, i.e. a link providing more details on the operator.

delegation-approver-address[optional]:
  Address to verify signatures when a staker wishes to delegate to the operator, as well as controlling "forced undelegations".
  Signature verification follows these rules:
  1) If this address is left as address(0), then any staker will be free to delegate to the operator, i.e. no signature verification will be performed.
  2) If this address is an EOA (i.e. it has no code), then we follow standard ECDSA signature verification for delegations to the operator.
  3) If this address is a contract (i.e. it has code) then we forward a call to the contract and verify that it returns the correct EIP-1271 "magic value".

pelldvs client operator modify-operator-details \
	--from <key-name> \
	--rpc-url <rpc-url> \
	--metadata-uri <metadata-uri> \
	--delegation-manager <delegation-manager> \
	<staker-optout-window-seconds> <delegation-approver-address>
`,
	Example: `
pelldvs client operator modify-operator-details \
	--from pell-localnet-deployer \
	--rpc-url http://localhost:8545 \
	--metadata-uri https://raw.githubusercontent.com/example/repo/file.json \
	--delegation-manager 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f \
	8600 0xa0Ee7A142d267C1f36714E4a8F75612F20a79720

pelldvs client operator modify-operator-details \
	--from pell-localnet-deployer \
	--rpc-url http://localhost:8545 \
	--delegation-manager 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f \
	8600

then you can query the operator details to see the changes:
pelldvs query operator operator-details \
	--rpc-url http://localhost:8545 \
	--delegation-manager 0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f \
	0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			// append empty string to args to make sure the length is 2
			args = append(args, chainutils.ZeroAddrStr)
		}
		return handleModifyOperatorDetails(cmd, args[0], args[1])
	},
}

func init() {
	chainflags.MetadataURI.AddToCmdFlag(modifyOperatorDetailsCmd)
}

func handleModifyOperatorDetails(
	cmd *cobra.Command,
	stakerOptoutWindowSeconds string,
	delegationApproverAddress string,
) error {
	optSeconds, err := chainutils.ConvStrToUint32(stakerOptoutWindowSeconds)
	if err != nil {
		return fmt.Errorf("failed to convert staker optout window seconds to uint32: %v", err)
	}

	if !gethcommon.IsHexAddress(delegationApproverAddress) {
		return fmt.Errorf("invalid delegation approver address %s", delegationApproverAddress)
	}

	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	_, err = ecdsa.ReadKey(kpath.ECDSA, "")
	if err != nil {
		return fmt.Errorf("failed to read ecdsa key: %v", err)
	}

	operator := types.Operator{
		DelegationApproverAddress: delegationApproverAddress,
		StakerOptOutWindow:        optSeconds,
	}
	if chainflags.MetadataURI.Value != "" {
		operator.MetadataURL = chainflags.MetadataURI.Value
	}

	receipt, err := execModifyOperatorDetails(cmd, kpath.ECDSA, operator)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "txHash", receipt.TxHash.String())

	return err
}

func execModifyOperatorDetails(cmd *cobra.Command, privKeyPath string, operator types.Operator) (*gethtypes.Receipt, error) {
	logger.Info(
		utils.GetPrettyCommandName(cmd),
		"privKeyPath", privKeyPath,
		"operator", operator,
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
	if !chainConfigChecker.IsValidPellDelegationManager() {
		return nil, fmt.Errorf("pell delegation manager is required")
	}

	chainOp, err := utils.NewOperatorFromCfg(cfg, logger)
	if err != nil {
		logger.Error("failed to create chain operator",
			"err", err,
			"file", pellcfg.CmtConfig.Pell.InteractorConfigPath,
			"cfg", fmt.Sprintf("%+v", cfg),
		)
		return nil, err
	}

	// get operator details first
	operatorDetails, err := chainOp.GetOperatorDetails(&bind.CallOpts{Context: ctx}, address.String())
	if err != nil {
		return nil, err
	}

	// modify operator details, update only the two fields
	operatorDetails.DelegationApproverAddress = operator.DelegationApproverAddress
	operatorDetails.StakerOptOutWindow = operator.StakerOptOutWindow
	if operator.MetadataURL != "" {
		operatorDetails.MetadataURL = operator.MetadataURL
	}

	receipt, err := chainOp.ModifyOperatorDetails(ctx, operatorDetails)

	logger.Info(
		utils.GetPrettyCommandName(cmd),
		"address", address,
		"receipt", receipt,
	)

	return receipt, err
}

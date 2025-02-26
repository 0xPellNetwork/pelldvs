package dvs

import (
	"context"
	"fmt"

	gethbind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/spf13/cobra"

	chaintypes "github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
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
	Long:  `query-dvs-info`,
	Example: `

pelldvs client dvs query-dvs-info --from pell-localnet-deployer --chain_id
pelldvs client dvs query-dvs-info --from pell-localnet-deployer --chain_id 666

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

	dvsInfo, err := execQueryDVSInfo(cmd, kpath.ECDSA, chainflags.ChainIDFlag.Value)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	fmt.Printf("dvsInfo : %+v", dvsInfo)

	return err
}

func execQueryDVSInfo(cmd *cobra.Command, privKeyPath string, chainID int) (*chaintypes.DVSInfo, error) {
	cmdName := "handleQueryDVSInfo"

	logger.Info(cmdName,
		"privKeyPath", privKeyPath,
		"chainID", chainID,
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

	dvsInfo, err := chainDVS.QueryDVSInfo(&gethbind.CallOpts{Context: ctx}, uint64(chainID))

	logger.Info(cmdName,
		"k", "v",
		"sender", senderAddress,
		"res", dvsInfo,
	)

	return dvsInfo, err
}

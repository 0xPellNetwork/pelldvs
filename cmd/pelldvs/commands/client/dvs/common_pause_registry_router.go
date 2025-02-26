package dvs

import (
	"fmt"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/chains/chainflags"
	"github.com/0xPellNetwork/pelldvs/cmd/pelldvs/commands/client/utils"
	pellcfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/pkg/keys"
)

func handlePauseOrUnRegistryRouter(cmd *cobra.Command, ok bool) error {
	kpath := keys.GetKeysPath(pellcfg.CmtConfig, chainflags.FromKeyNameFlag.Value)
	if !kpath.IsECDSAExist() {
		return fmt.Errorf("ECDSA key does not exist %s", kpath.ECDSA)
	}

	res, err := execPauseOrRegistryRouter(cmd, kpath.ECDSA, ok)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	logger.Info("tx successfully", "res", res.TxHash.String())

	return err
}

func execPauseOrRegistryRouter(cmd *cobra.Command, privKeyPath string, ok bool) (*gethtypes.Receipt, error) {
	cmdName := "handlePauseRegistryRouter"

	logger.Info(cmdName,
		"privKeyPath", privKeyPath,
	)

	ctx := cmd.Context()
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
	var receipt *gethtypes.Receipt
	if ok {
		receipt, err = chainDVS.PauseRegistryRouter(ctx)
	} else {
		receipt, err = chainDVS.UnPauseRegistryRouter(ctx)
	}

	logger.Info(cmdName,
		"k", "v",
		"sender", senderAddress,
		"receipt", receipt,
	)

	return receipt, err
}

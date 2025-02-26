package utils

import (
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-interactor/chainlibs/eth"
	chaincfg "github.com/0xPellNetwork/pelldvs-interactor/config"
	"github.com/0xPellNetwork/pelldvs-interactor/interactor/operator"
	"github.com/0xPellNetwork/pelldvs-libs/log"
)

func NewOperatorFromFile(cmd *cobra.Command, filepath string, logger log.Logger) (*operator.Operator, *chaincfg.Config, error) {
	cfg, err := LoadChainConfig(cmd, filepath, logger)
	if err != nil {
		logger.Error("failed to load chain config",
			"file", filepath,
			"error", err,
		)
		return nil, nil, err
	}

	chainOp, err := NewOperatorFromCfg(cfg, logger)
	return chainOp, cfg, err
}

func NewOperatorFromCfg(cfg *chaincfg.Config, logger log.Logger) (*operator.Operator, error) {
	client, err := eth.NewClient(cfg.RPCURL)
	if err != nil {
		return nil, err
	}

	chainOp, err := operator.New(cfg, client, logger)
	if err != nil {
		return nil, err
	}

	return chainOp, nil
}

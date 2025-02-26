package utils

import (
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pelldvs-interactor/chainlibs/eth"
	chaincfg "github.com/0xPellNetwork/pelldvs-interactor/config"
	chaindvs "github.com/0xPellNetwork/pelldvs-interactor/interactor/dvs"
	"github.com/0xPellNetwork/pelldvs-libs/log"
)

func NewDVSFromFromFile(cmd *cobra.Command, filepath string, logger log.Logger) (*chaindvs.DVS, *chaincfg.Config, error) {
	cfg, err := LoadChainConfig(cmd, filepath, logger)
	if err != nil {
		logger.Error("failed to load chain config", "error", err)
		return nil, nil, err
	}

	chainDVS, err := NewDVSFromCfg(cfg, logger)
	return chainDVS, cfg, err
}

func NewDVSFromCfg(cfg *chaincfg.Config, logger log.Logger) (*chaindvs.DVS, error) {
	client, err := eth.NewClient(cfg.RPCURL)
	if err != nil {
		return nil, err
	}
	dvs, err := chaindvs.New(cfg, client, logger)
	if err != nil {
		return nil, err
	}
	return dvs, nil
}

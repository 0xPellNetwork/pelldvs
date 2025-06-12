package utils

import (
	"context"
	"fmt"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/ethereum/go-ethereum/common"

	interactorcfg "github.com/0xPellNetwork/pelldvs-interactor/config"
	"github.com/0xPellNetwork/pelldvs-interactor/interactor/indexer"
	"github.com/0xPellNetwork/pelldvs-interactor/interactor/reader"
	"github.com/0xPellNetwork/pelldvs-interactor/libs/chainregistry"
	"github.com/0xPellNetwork/pelldvs-libs/log"
)

func CreateDVSReader(
	ctx context.Context,
	interactorConfig *interactorcfg.Config,
	db dbm.DB,
	logger log.Logger,
) (reader.DVSReader, error) {
	registry, err := chainregistry.NewRegistry(ctx, logger, chainregistry.WithInteractorConfig(interactorConfig))
	if err != nil {
		logger.Error("Failed to create chain registry", "error", err)
		return nil, fmt.Errorf("failed to create chain registry: %v", err)
	}
	logger.Info("Chain registry", "registry", registry)

	pellIndexerConfig := indexer.NewPellIndexerConfig(
		interactorConfig.ContractConfig.IndexerStartHeight,
		interactorConfig.ContractConfig.IndexerBatchSize,
		interactorConfig.ContractConfig.IndexerListenInterval,
		common.HexToAddress(interactorConfig.ContractConfig.PellDelegationManager),
		common.HexToAddress(interactorConfig.ContractConfig.PellRegistryRouter),
	)
	pellindexer, err := indexer.NewInitedPellIndexer(ctx, registry, interactorConfig.ChainID, pellIndexerConfig, db, logger)
	if err != nil {
		logger.Error("Failed to create Pell indexer", "error", err)
		return nil, fmt.Errorf("failed to create Pell indexer: %v", err)
	}

	dvsReaderConfig := reader.NewDVSReaderConfig(interactorConfig)
	dvsReader, err := reader.NewDVSReader(ctx, pellindexer, dvsReaderConfig, logger)
	if err != nil {
		logger.Error("Failed to create DVS reader", "error", err)
		return nil, fmt.Errorf("failed to create DVS reader: %v", err)
	}

	return dvsReader, nil
}

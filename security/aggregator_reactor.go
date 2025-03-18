package security

import (
	"github.com/0xPellNetwork/pelldvs-libs/log"
	aggtypes "github.com/0xPellNetwork/pelldvs/aggregator"
	"github.com/0xPellNetwork/pelldvs/types"
)

type AggregatorReactor struct {
	aggregator   aggtypes.Aggregator
	logger       log.Logger
	eventManager *EventManager
}

func CreateAggregatorReactor(
	aggregator aggtypes.Aggregator,
	logger log.Logger,
	eventManager *EventManager,
) *AggregatorReactor {
	return &AggregatorReactor{
		aggregator:   aggregator,
		logger:       logger,
		eventManager: eventManager,
	}
}

func (ar *AggregatorReactor) HandleSignatureCollectionRequest() {
	ar.eventManager.eventBus.Publish(types.CollectResponseSignatureDone)
}

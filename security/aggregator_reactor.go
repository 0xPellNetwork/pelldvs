package security

import (
	"github.com/0xPellNetwork/pelldvs-libs/log"
	aggtypes "github.com/0xPellNetwork/pelldvs/aggregator"
	"github.com/0xPellNetwork/pelldvs/types"
)

type AggregatorReactor struct {
	aggregator aggtypes.Aggregator
	logger     log.Logger
	bus        *types.ReactorEventBus
}

func CreateAggregatorReactor(
	aggregator aggtypes.Aggregator,
	logger log.Logger,
	bus *types.ReactorEventBus,
) *AggregatorReactor {

	return &AggregatorReactor{
		aggregator: aggregator,
		logger:     logger,
		bus:        bus,
	}
}

func (ar *AggregatorReactor) HandleSignatureCollectionRequest() {
	ar.bus.Publish(types.CollectResponseSignatureDone)
}

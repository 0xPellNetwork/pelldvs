package security

import (
	"github.com/0xPellNetwork/pelldvs-libs/log"
	avsitypes "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/types"
)

type EventManager struct {
	eventBus          *types.ReactorEventBus
	aggregatorReactor *AggregatorReactor
	dvsReactor        *DVSReactor
	logger            log.Logger
}

func NewEventManager(logger log.Logger) *EventManager {
	return &EventManager{
		logger:   logger,
		eventBus: types.NewReactorEventBus(),
	}
}

func (em *EventManager) SetDVSReactor(dvsReactor *DVSReactor) {
	em.dvsReactor = dvsReactor
}

func (em *EventManager) SetAggregatorReactor(aggregatorReactor *AggregatorReactor) {
	em.aggregatorReactor = aggregatorReactor
}

// StartListening starts listening to the event bus
func (em *EventManager) StartListening() {
	go func() {
		requestCh := em.eventBus.Subscribe(types.CollectResponseSignatureRequest)
		responseCh := em.eventBus.Subscribe(types.CollectResponseSignatureDone)

		for {
			select {
			case event := <-requestCh:
				// Handle CollectResponseSignatureRequest
				if event.Type == types.CollectResponseSignatureRequest {
					em.logger.Info("Received CollectResponseSignatureRequest")

					requestHash := event.Payload.(avsitypes.DVSRequestHash)
					if err := em.aggregatorReactor.HandleSignatureCollectionRequest(requestHash); err != nil {
						em.logger.Error("failed to handle aggregator request", "error", err)
					}

					em.logger.Info("Handled CollectResponseSignatureRequest")
				}

			case event := <-responseCh:
				// Handle CollectResponseSignatureDone
				if event.Type == types.CollectResponseSignatureDone {
					em.logger.Info("Received CollectResponseSignatureDone")

					res := event.Payload.(AggregatorResponse)
					err := em.dvsReactor.OnRequestAfterAggregated(res.requestHash, res.validateResponse)
					if err != nil {
						em.logger.Error("failed to handle aggregator request", "error", err)
					}

					em.logger.Info("Handled CollectResponseSignatureDone")
				}
			}
		}
	}()
}

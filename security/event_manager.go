package security

import (
	"fmt"

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

func (em *EventManager) StartListening() {
	go func() {
		requestCh := em.eventBus.Subscribe(types.CollectResponseSignatureRequest)
		responseCh := em.eventBus.Subscribe(types.CollectResponseSignatureDone)

		for {
			select {
			case event := <-requestCh:
				if event.Type == types.CollectResponseSignatureRequest {
					em.logger.Info(fmt.Sprintf("Received CollectResponseSignatureRequest"))

					requestHash := event.Payload.(avsitypes.DVSRequestHash)
					err := em.aggregatorReactor.HandleSignatureCollectionRequest(requestHash)
					if err != nil {
						em.logger.Error("failed to handle aggregator request: %v", err)
					}

					em.logger.Info(fmt.Sprintf("Handled CollectResponseSignatureRequest"))
				}

			case event := <-responseCh:
				if event.Type == types.CollectResponseSignatureDone {
					em.logger.Info(fmt.Sprintf("Received CollectResponseSignatureDone"))

					res := event.Payload.(AggregatorResponse)
					err := em.dvsReactor.OnRequestAfterAggregated(res.requestHash, res.validateResponse)
					if err != nil {
						fmt.Println("[EventManager] Error notifying DVS: ", err)
					}

					em.logger.Info(fmt.Sprintf("Handled CollectResponseSignatureDone"))
				}
			}
		}
	}()
}

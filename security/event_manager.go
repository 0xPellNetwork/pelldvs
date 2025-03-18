package security

import (
	"fmt"
	"github.com/0xPellNetwork/pelldvs/types"
)

type EventManager struct {
	eventBus          *types.ReactorEventBus
	aggregatorReactor *AggregatorReactor
	dvsReactor        *DVSReactor
}

func NewEventManager() *EventManager {
	return &EventManager{
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
				if event == types.CollectResponseSignatureRequest {
					fmt.Println("[EventManager] Received CollectResponseSignatureRequest, forwarding to Aggregator")
					em.aggregatorReactor.HandleSignatureCollectionRequest()
				}
			case event := <-responseCh:
				if event == types.CollectResponseSignatureDone {
					fmt.Println("[EventManager] Received CollectResponseSignatureDone, notifying DVS")
					em.dvsReactor.HandleSignatureCollectionResponse()
				}
			}
		}
	}()
}

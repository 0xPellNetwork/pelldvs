package security

import (
	"fmt"
	avsitypes "github.com/0xPellNetwork/pelldvs/avsi/types"
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
				if event.Type == types.CollectResponseSignatureRequest {
					fmt.Println("[EventManager] Received CollectResponseSignatureRequest, forwarding to Aggregator")
					requestHash := event.Payload.(avsitypes.DVSRequestHash)
					err := em.aggregatorReactor.HandleSignatureCollectionRequest(requestHash)
					if err != nil {
						fmt.Println("[EventManager] Error handling signature collection request: ", err)
					} else {
						fmt.Println("[EventManager] Handled signature collection request")
					}
				}
			case event := <-responseCh:
				if event.Type == types.CollectResponseSignatureDone {
					fmt.Println("[EventManager] Received CollectResponseSignatureDone, notifying DVS")
					res := event.Payload.(AggregatorResponse)
					err := em.dvsReactor.OnRequestAfterAggregated(res.requestHash, res.validateResponse)
					if err != nil {
						fmt.Println("[EventManager] Error notifying DVS: ", err)
					} else {
						fmt.Println("[EventManager] Notified DVS with response")
					}
				}
			}
		}
	}()
}

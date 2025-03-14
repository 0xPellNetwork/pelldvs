package security

import (
	"fmt"
	"github.com/0xPellNetwork/pelldvs/types"
)

type EventManager struct {
	bus        *types.ReactorEventBus
	aggregator *AggregatorReactor
	dvs        *AggregatorReactor
}

func NewEventManager(
	bus *types.ReactorEventBus,
	aggregator *AggregatorReactor,
	dvs *AggregatorReactor,
) *EventManager {

	return &EventManager{
		bus:        bus,
		aggregator: aggregator,
		dvs:        dvs,
	}
}

func (em *EventManager) StartListening() {
	go func() {
		requestCh := em.bus.Subscribe(types.CollectResponseSignatureRequest)
		responseCh := em.bus.Subscribe(types.CollectResponseSignatureDone)

		for {
			select {
			case event := <-requestCh:
				if event == types.CollectResponseSignatureRequest {
					fmt.Println("[EventManager] Received CollectResponseSignatureRequest, forwarding to Aggregator")
					em.aggregator.HandleSignatureCollectionRequest()
				}
			case event := <-responseCh:
				if event == types.CollectResponseSignatureDone {
					fmt.Println("[EventManager] Received CollectResponseSignatureDone, notifying DVS")
					em.dvs.HandleSignatureCollectionRequest()
				}
			}
		}
	}()
}

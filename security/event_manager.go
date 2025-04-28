package security

import (
	"github.com/0xPellNetwork/pelldvs-libs/log"
	avsitypes "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/types"
)

// EventManager coordinates communication between different components of the system
// by managing event subscriptions and routing events to appropriate handlers
type EventManager struct {
	logger            log.Logger
	eventBus          *types.ReactorEventBus
	aggregatorReactor *AggregatorReactor
	dvsReactor        *DVSReactor
}

// NewEventManager creates a new EventManager instance with the provided logger
// and initializes an internal event bus for communication
func NewEventManager(logger log.Logger) *EventManager {
	return &EventManager{
		logger:   logger,
		eventBus: types.NewReactorEventBus(),
	}
}

// SetDVSReactor assigns the DVS reactor component to the EventManager
// allowing it to forward relevant events to the DVS subsystem
func (em *EventManager) SetDVSReactor(dvsReactor *DVSReactor) {
	em.dvsReactor = dvsReactor
}

// SetAggregatorReactor assigns the aggregator reactor component to the EventManager
// allowing it to forward signature collection events to the aggregation subsystem
func (em *EventManager) SetAggregatorReactor(aggregatorReactor *AggregatorReactor) {
	em.aggregatorReactor = aggregatorReactor
}

// StartListening begins asynchronous event processing by subscribing to relevant
// event types and dispatching them to the appropriate handlers
func (em *EventManager) StartListening() {
	go func() {
		// Subscribe to signature collection request and completion events
		requestCh := em.eventBus.Sub(types.CollectResponseSignatureRequest)
		responseCh := em.eventBus.Sub(types.CollectResponseSignatureDone)

		for {
			select {
			case event := <-requestCh:
				// Handle signature collection request events
				if event.Type == types.CollectResponseSignatureRequest {
					em.logger.Info("Received CollectResponseSignatureRequest")

					// Extract the request hash from the event payload
					requestHash := event.Payload.(avsitypes.DVSRequestHash)
					// Forward to the aggregator reactor for processing
					if err := em.aggregatorReactor.HandleSignatureCollectionRequest(requestHash); err != nil {
						em.logger.Error("failed to handle aggregator request", "error", err)
					}

					em.logger.Info("Handled CollectResponseSignatureRequest")
				}

			case event := <-responseCh:
				// Handle signature collection completion events
				if event.Type == types.CollectResponseSignatureDone {
					em.logger.Info("Received CollectResponseSignatureDone")

					// Extract the aggregated response from the event payload
					res := event.Payload.(AggregatorResponse)
					// Forward to the DVS reactor for post-aggregation processing
					if err := em.dvsReactor.OnRequestAfterAggregated(res.requestHash, res.validateResponse); err != nil {
						em.logger.Error("failed to handle aggregator request", "error", err)
					}

					em.logger.Info("Handled CollectResponseSignatureDone")
				}
			}
		}
	}()
}

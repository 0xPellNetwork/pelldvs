package types

import (
	"sync"
)

type ReactorEventType string

const (
	CollectResponseSignatureRequest ReactorEventType = "CollectResponseSignatureRequest"
	CollectResponseSignatureDone    ReactorEventType = "CollectResponseSignatureDone"
)

// ReactorEvent is a struct that represents an event that can be published to the ReactorEventBus.
type ReactorEvent struct {
	Type    ReactorEventType
	Payload any
}

// ReactorEventBus is a struct that manages event subscriptions and publications.
type ReactorEventBus struct {
	channels sync.Map // map[ReactorEventType]chan ReactorEvent
}

// NewReactorEventBus creates a new ReactorEventBus.
func NewReactorEventBus() *ReactorEventBus {
	return &ReactorEventBus{
		channels: sync.Map{},
	}
}

// Sub subscribes to a ReactorEventType and returns a channel that will receive ReactorEvents.
func (eb *ReactorEventBus) Sub(eventType ReactorEventType) <-chan ReactorEvent {
	ch := make(chan ReactorEvent, 16)
	eb.channels.Store(eventType, ch)
	return ch
}

// Pub publishes a ReactorEvent to all subscribers of the given ReactorEventType.
func (eb *ReactorEventBus) Pub(eventType ReactorEventType, payload any) {
	if chAny, exists := eb.channels.Load(eventType); exists {
		ch := chAny.(chan ReactorEvent)
		ch <- ReactorEvent{
			Type:    eventType,
			Payload: payload,
		}
	}
}

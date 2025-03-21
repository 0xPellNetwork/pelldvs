package types

import (
	"sync"
)

type ReactorEventBus struct {
	channels map[ReactorEventType]chan ReactorEvent
	mu       sync.RWMutex
}

func NewReactorEventBus() *ReactorEventBus {
	return &ReactorEventBus{
		channels: make(map[ReactorEventType]chan ReactorEvent),
	}
}

func (eb *ReactorEventBus) Subscribe(eventType ReactorEventType) <-chan ReactorEvent {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	ch := make(chan ReactorEvent, 16)
	eb.channels[eventType] = ch
	return ch
}

func (eb *ReactorEventBus) Publish(eventType ReactorEventType, payload any) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	if ch, exists := eb.channels[eventType]; exists {
		ch <- ReactorEvent{
			Type:    eventType,
			Payload: payload,
		}
	}
}

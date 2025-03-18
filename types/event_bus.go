package types

import (
	"sync"
)

type ReactorEventBus struct {
	channels map[ReactorEvent]chan ReactorEvent
	mu       sync.RWMutex
}

// NewReactorEventBus returns a new event bus.
func NewReactorEventBus() *ReactorEventBus {
	return &ReactorEventBus{
		channels: make(map[ReactorEvent]chan ReactorEvent),
	}
}

func (eb *ReactorEventBus) Subscribe(event ReactorEvent) <-chan ReactorEvent {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	ch := make(chan ReactorEvent, 16)
	eb.channels[event] = ch
	return ch
}

func (eb *ReactorEventBus) Publish(event ReactorEvent) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	if ch, exists := eb.channels[event]; exists {
		ch <- event
	}
}

package types

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

// TestNewReactorEventBus tests the creation of a new ReactorEventBus.
func TestNewReactorEventBus(t *testing.T) {
	eventBus := NewReactorEventBus()
	assert.NotNil(t, eventBus)
	assert.IsType(t, &sync.Map{}, &eventBus.channels)
}

// TestSubscribe tests the subscription to a ReactorEventType.
func TestSubscribe(t *testing.T) {
	eventBus := NewReactorEventBus()
	eventType := ReactorEventType("test_event")
	ch := eventBus.Subscribe(eventType)
	assert.NotNil(t, ch)
}

// TestPublish tests the publication of a ReactorEvent to subscribers.
func TestPublish(t *testing.T) {
	eventBus := NewReactorEventBus()
	eventType := ReactorEventType("test_event")
	ch := eventBus.Subscribe(eventType)

	payload := "test_payload"
	eventBus.Publish(eventType, payload)

	event := <-ch
	assert.Equal(t, eventType, event.Type)
	assert.Equal(t, payload, event.Payload)
}

// TestPublishWithoutSubscribers tests the publication of a ReactorEvent without subscribers.
func TestPublishWithoutSubscribers(t *testing.T) {
	eventBus := NewReactorEventBus()
	eventType := ReactorEventType("test_event")

	// Publish without any subscribers
	eventBus.Publish(eventType, "test_payload")

	// Ensure no panic or error occurs
	assert.True(t, true)
}

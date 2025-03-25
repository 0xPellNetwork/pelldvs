package types

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

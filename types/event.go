package types

type ReactorEventType string

const (
	CollectResponseSignatureRequest ReactorEventType = "CollectResponseSignatureRequest"
	CollectResponseSignatureDone    ReactorEventType = "CollectResponseSignatureDone"
)

type ReactorEvent struct {
	Type    ReactorEventType
	Payload any
}

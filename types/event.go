package types

type ReactorEvent string

const (
	CollectResponseSignatureRequest ReactorEvent = "CollectResponseSignatureRequest"
	CollectResponseSignatureDone    ReactorEvent = "CollectResponseSignatureDone"
)

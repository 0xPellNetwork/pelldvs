package types

import (
	"crypto/sha256"
)

// DvsRequestKeySize is the size of the DVS request key index
const DvsRequestKeySize = sha256.Size

const (
	// EventTypeKey is a reserved composite key for event name.
	EventTypeKey = "dvs.event"
	// DVSHashKey is a reserved key, used to specify DVS request's hash.
	DVSHashKey = "dvs.hash"

	// DVSHeightKey is a reserved key, used to specify DVS request block's height.
	DVSHeightKey = "dvs.height"
	DVSChainID   = "dvs.chainid"
)

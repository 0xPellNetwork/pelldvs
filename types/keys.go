package types

import "crypto/sha256"

// UNSTABLE
// PeerStateKey represents the key used to store peer state information
// in the consensus reactor's data structures
var (
	PeerStateKey = "ConsensusReactor.peerState"
)

// DvsRequestKeySize is the size of the DVS request key index
// Using SHA-256 hash size as the standard key size for indexing requests
const DvsRequestKeySize = sha256.Size

// Event and index key constants used throughout the system
// for consistent event handling and data retrieval
const (
	// EventTypeKey is a reserved composite key for event name.
	// Used as a standard key in event attribute maps to identify event types
	EventTypeKey = "dvs.event"

	// DVSHashKey is a reserved key, used to specify DVS request's hash.
	// Critical for indexing and retrieving DVS requests by their unique hash
	DVSHashKey = "dvs.hash"

	// DVSHeightKey is a reserved key, used to specify DVS request block's height.
	// Enables filtering and retrieving DVS requests by blockchain height
	DVSHeightKey = "dvs.height"

	// DVSChainID identifies which blockchain the DVS request belongs to
	// Important for multi-chain environments to route requests properly
	DVSChainID = "dvs.chainid"
)

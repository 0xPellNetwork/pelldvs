package types

import (
	"crypto/sha256"
	"fmt"

	"github.com/0xPellNetwork/pelldvs/crypto/tmhash"
)

// DvsRequestKeySize is the size of the DVS request key index
const DvsRequestKeySize = sha256.Size

const (
	// EventTypeKey is a reserved composite key for event name.
	EventTypeKey = "dvs.event"
	// DVSHashKey is a reserved key, used to specify DVS request's hash.
	// see EventBus#PublishEventTx
	DVSHashKey = "dvs.hash"

	// DVSHeightKey is a reserved key, used to specify DVS request block's height.
	DVSHeightKey = "dvs.height"
	DVSChainID   = "dvs.chainid"
)

type (
	DvsRequest []byte
	// TxKey is the fixed length array key used as an index.
	DvsRequestKey [DvsRequestKeySize]byte
)

// Hash computes the TMHASH hash of the wire encoded DVS request.
func (dvs DvsRequest) Hash() []byte {
	return tmhash.Sum(dvs)
}

func (dvs DvsRequest) Key() DvsRequestKey {
	return sha256.Sum256(dvs)
}

// String returns the hex-encoded DVS request as a string.
func (dvs DvsRequest) String() string {
	return fmt.Sprintf("Tx{%X}", []byte(dvs))
}

package kv

import (
	"fmt"

	"github.com/google/orderedcode"
)

func ParseEventSeqFromEventKey(key []byte) (int64, error) {
	var (
		compositeKey, typ, eventValue string
		height                        int64
		eventSeq                      int64
	)

	remaining, err := orderedcode.Parse(string(key), &compositeKey, &eventValue, &height, &typ, &eventSeq)
	if err != nil {
		return 0, fmt.Errorf("failed to parse event key: %w", err)
	}

	if len(remaining) != 0 {
		return 0, fmt.Errorf("unexpected remainder in key: %s", remaining)
	}

	return eventSeq, nil
}

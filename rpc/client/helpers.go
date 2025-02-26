package client

import (
	"fmt"
	"time"
)

// Waiter is informed of current height, decided whether to quit early
type Waiter func(delta int64) (abort error)

// DefaultWaitStrategy is the standard backoff algorithm,
// but you can plug in another one
func DefaultWaitStrategy(delta int64) (abort error) {
	if delta > 10 {
		return fmt.Errorf("waiting for %d blocks... aborting", delta)
	} else if delta > 0 {
		// estimate of wait time....
		// wait half a second for the next block (in progress)
		// plus one second for every full block
		delay := time.Duration(delta-1)*time.Second + 500*time.Millisecond
		time.Sleep(delay)
	}
	return nil
}

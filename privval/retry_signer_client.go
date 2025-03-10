package privval

import (
	"time"

	"github.com/0xPellNetwork/pelldvs/crypto/bls"
)

// RetrySignerClient wraps SignerClient adding retry for each operation (except
// Ping) w/ a timeout.
type RetrySignerClient struct {
	next    *SignerClient
	retries int
	timeout time.Duration
}

// NewRetrySignerClient returns RetrySignerClient. If +retries+ is 0, the
// client will be retrying each operation indefinitely.
func NewRetrySignerClient(sc *SignerClient, retries int, timeout time.Duration) *RetrySignerClient {
	return &RetrySignerClient{sc, retries, timeout}
}

// var _ types.PrivValidator = (*RetrySignerClient)(nil)

func (sc *RetrySignerClient) Close() error {
	return sc.next.Close()
}

func (sc *RetrySignerClient) IsConnected() bool {
	return sc.next.IsConnected()
}

func (sc *RetrySignerClient) WaitForConnection(maxWait time.Duration) error {
	return sc.next.WaitForConnection(maxWait)
}

//--------------------------------------------------------
// Implement PrivValidator

func (sc *RetrySignerClient) Ping() error {
	return sc.next.Ping()
}

func (sc *RetrySignerClient) GetPubKey() (*bls.G1Point, error) {
	// TODO implement me
	panic("implement me")
}

func (sc *RetrySignerClient) SignBytes(bytes []byte) (*bls.Signature, error) {
	//TODO implement me
	panic("implement me")
}

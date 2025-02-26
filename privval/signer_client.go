package privval

import (
	"fmt"
	"time"

	"github.com/0xPellNetwork/pelldvs/crypto"
	cryptoenc "github.com/0xPellNetwork/pelldvs/crypto/encoding"
	privvalproto "github.com/0xPellNetwork/pelldvs/proto/pelldvs/privval"
)

// SignerClient implements PrivValidator.
// Handles remote validator connections that provide signing services
type SignerClient struct {
	endpoint *SignerListenerEndpoint
	chainID  string
}

// var _ types.PrivValidator = (*SignerClient)(nil)

// NewSignerClient returns an instance of SignerClient.
// it will start the endpoint (if not already started)
func NewSignerClient(endpoint *SignerListenerEndpoint, chainID string) (*SignerClient, error) {
	if !endpoint.IsRunning() {
		if err := endpoint.Start(); err != nil {
			return nil, fmt.Errorf("failed to start listener endpoint: %w", err)
		}
	}

	return &SignerClient{endpoint: endpoint, chainID: chainID}, nil
}

// Close closes the underlying connection
func (sc *SignerClient) Close() error {
	return sc.endpoint.Close()
}

// IsConnected indicates with the signer is connected to a remote signing service
func (sc *SignerClient) IsConnected() bool {
	return sc.endpoint.IsConnected()
}

// WaitForConnection waits maxWait for a connection or returns a timeout error
func (sc *SignerClient) WaitForConnection(maxWait time.Duration) error {
	return sc.endpoint.WaitForConnection(maxWait)
}

//--------------------------------------------------------
// Implement PrivValidator

// Ping sends a ping request to the remote signer
func (sc *SignerClient) Ping() error {
	response, err := sc.endpoint.SendRequest(mustWrapMsg(&privvalproto.PingRequest{}))
	if err != nil {
		sc.endpoint.Logger.Error("SignerClient::Ping", "err", err)
		return nil
	}

	pb := response.GetPingResponse()
	if pb == nil {
		return err
	}

	return nil
}

// GetPubKey retrieves a public key from a remote signer
// returns an error if client is not able to provide the key
func (sc *SignerClient) GetPubKey() (crypto.PubKey, error) {
	response, err := sc.endpoint.SendRequest(mustWrapMsg(&privvalproto.PubKeyRequest{ChainId: sc.chainID}))
	if err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	resp := response.GetPubKeyResponse()
	if resp == nil {
		return nil, ErrUnexpectedResponse
	}
	if resp.Error != nil {
		return nil, &RemoteSignerError{Code: int(resp.Error.Code), Description: resp.Error.Description}
	}

	pk, err := cryptoenc.PubKeyFromProto(resp.PubKey)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

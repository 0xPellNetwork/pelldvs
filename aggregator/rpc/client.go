package rpc

import (
	"fmt"
	"net/rpc"

	aggTypes "github.com/0xPellNetwork/pelldvs/aggregator"
)

// RPCClientAggregator provides a client implementation of the Aggregator interface
// communicating with the aggregator service over RPC
type RPCClientAggregator struct {
	client *rpc.Client
}

// NewRPCClientAggregator creates a new client instance connected to the specified address
// establishing a connection to the aggregator RPC server
func NewRPCClientAggregator(address string) (*RPCClientAggregator, error) {
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to aggregator: %v", err)
	}
	return &RPCClientAggregator{
		client: client,
	}, nil
}

// CollectResponseSignature implements the Aggregator interface by forwarding the request
// to the RPC server and receiving the validated response
func (ra *RPCClientAggregator) CollectResponseSignature(response *aggTypes.ResponseWithSignature, validatedResponseCh chan<- aggTypes.ValidatedResponse) error {
	var result aggTypes.ValidatedResponse
	err := ra.client.Call("RPCServerAggregator.CollectResponseSignature", response, &result)
	if err != nil {
		return fmt.Errorf("failed to call aggregator RPC method: %v", err)
	}

	validatedResponseCh <- result
	return nil
}

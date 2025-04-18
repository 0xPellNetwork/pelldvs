package rpc

import (
	"fmt"
	"net/rpc"

	aggTypes "github.com/0xPellNetwork/pelldvs/aggregator"
)

// RPCClientAggregator is an implementation that interacts with the aggregator via RPC
type RPCClientAggregator struct {
	client *rpc.Client
}

// NewRPCClientAggregator creates a new instance of RPCClientAggregator
func NewRPCClientAggregator(address string) (*RPCClientAggregator, error) {
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to aggregator: %v", err)
	}
	return &RPCClientAggregator{
		client: client,
	}, nil
}

// CollectResponseSignature implements the CollectResponseSignature method of the Aggregator interface
func (ra *RPCClientAggregator) CollectResponseSignature(response *aggTypes.ResponseWithSignature, validatedResponseCh chan<- aggTypes.ValidatedResponse) error {
	var result aggTypes.ValidatedResponse
	err := ra.client.Call("RPCServerAggregator.CollectResponseSignature", response, &result)
	if err != nil {
		return fmt.Errorf("failed to call aggregator RPC method: %v", err)
	}

	validatedResponseCh <- result
	return nil
}

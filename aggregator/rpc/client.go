package rpc

import (
	"errors"
	"fmt"
	"net/rpc"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	aggTypes "github.com/0xPellNetwork/pelldvs/aggregator"
	"github.com/0xPellNetwork/pelldvs/libs/clientmanager/rpcclientmanager"
)

const (
	RPCHealthCheckMethod      = "RPCServerAggregator.HealthCheck"
	RPCServerAggregatorMethod = "RPCServerAggregator.CollectResponseSignature"
)

// RPCClientAggregator provides a client implementation of the Aggregator interface
// communicating with the aggregator service over RPC
type RPCClientAggregator struct {
	clientManager *rpcclientmanager.RPCClientManager // Use RPCClientManager for connection management
	logger        log.Logger
}

// NewRPCClientAggregator creates a new client instance connected to the specified address
// establishing a connection to the aggregator RPC server
func NewRPCClientAggregator(address string, logger log.Logger) (*RPCClientAggregator, error) {
	manager, err := rpcclientmanager.NewRPCClientManager(address, RPCHealthCheckMethod, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to aggregator: %v", err)
	}
	return &RPCClientAggregator{
		clientManager: manager,
		logger:        logger,
	}, nil
}

// CollectResponseSignature implements the Aggregator interface by forwarding the request
// to the RPC server and receiving the validated response
func (ra *RPCClientAggregator) CollectResponseSignature(response *aggTypes.ResponseWithSignature, validatedResponseCh chan<- aggTypes.ValidatedResponse) error {
	var result aggTypes.ValidatedResponse
	client, err := ra.clientManager.GetClient()
	if err != nil {
		return fmt.Errorf("failed to get RPC client: %v", err)
	}
	err = client.Call(RPCServerAggregatorMethod, response, &result)

	if errors.Is(err, rpc.ErrShutdown) {
		// If the client is shutdown, try to reconnect
		ra.logger.Info("RPC client is shutdown, attempting to reconnect")
		client, err = ra.clientManager.GetClient()
		if err != nil {
			return fmt.Errorf("failed to get RPC client after shutdown: %v", err)
		}
		err = client.Call(RPCServerAggregatorMethod, response, &result)
	}

	if err != nil {
		return fmt.Errorf("failed to call aggregator RPC method: %v", err)
	}
	validatedResponseCh <- result
	return nil
}

// HealthCheck performs a health check on the aggregator service
func (ra *RPCClientAggregator) HealthCheck() (bool, error) {
	var result bool
	client, err := ra.clientManager.GetClient()
	if err != nil {
		return false, fmt.Errorf("failed to get RPC client: %v", err)
	}
	err = client.Call(RPCHealthCheckMethod, struct{}{}, &result)
	if err != nil {
		return false, fmt.Errorf("failed to call aggregator RPC method: %v", err)
	}
	return result, nil
}

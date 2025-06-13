package rpc

import (
	"context"
	"errors"
	"fmt"
	"net/rpc"

	"github.com/0xPellNetwork/pelldvs-interactor/libs/clientmanager/rpcclientmanager"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	aggTypes "github.com/0xPellNetwork/pelldvs/aggregator"
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
	manager, err := rpcclientmanager.NewRPCClientManager(
		address,
		logger,
		rpcclientmanager.WithHealthCheckMethod(RPCHealthCheckMethod),
	)
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
	ctx := context.Background()
	client, err := ra.clientManager.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get RPC client: %v", err)
	}
	ra.logger.Info("AggregatorClient: Calling RPC method to collect response signature", "response", response)
	err = client.Call(RPCServerAggregatorMethod, response, &result)

	if errors.Is(err, rpc.ErrShutdown) {
		// If the client is shutdown, try to reconnect
		ra.logger.Info("RPC client is shutdown, attempting to reconnect")
		client, err = ra.clientManager.GetClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to get RPC client after shutdown: %v", err)
		}
		err = client.Call(RPCServerAggregatorMethod, response, &result)
	}

	if err != nil {
		return fmt.Errorf("failed to call aggregator RPC method: %v", err)
	}

	if result.Err != nil {
		return fmt.Errorf("aggregator returned error: %v", result.Err)
	}

	ra.logger.Info("AggregatorClient: Received validated response", "result", result)
	validatedResponseCh <- result
	ra.logger.Info("AggregatorClient: Sent validated response to channel")
	return nil
}

// HealthCheck performs a health check on the aggregator service
func (ra *RPCClientAggregator) HealthCheck() (bool, error) {
	var result bool
	client, err := ra.clientManager.GetClient(context.Background())
	if err != nil {
		return false, fmt.Errorf("failed to get RPC client: %v", err)
	}

	if err = client.Call(RPCHealthCheckMethod, struct{}{}, &result); err != nil {
		return false, fmt.Errorf("failed to call aggregator RPC method: %v", err)
	}

	return result, nil
}

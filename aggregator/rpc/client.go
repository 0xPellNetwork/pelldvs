package rpc

import (
	"context"
	"errors"
	"fmt"
	"net/rpc"

	"github.com/0xPellNetwork/pelldvs-interactor/libs/clientmanager/rpcclientmanager"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	aggtypes "github.com/0xPellNetwork/pelldvs/aggregator/types"
)

const (
	RPCHealthCheckMethod      = "AggregatorRPCServer.HealthCheck"
	RPCServerAggregatorMethod = "AggregatorRPCServer.CollectResponseSignature"
)

// AggregatorRPCClient provides a client implementation of the Aggregator interface
// communicating with the aggregator service over RPC
type AggregatorRPCClient struct {
	clientManager *rpcclientmanager.RPCClientManager // Use RPCClientManager for connection management
	logger        log.Logger
}

// NewAggregatorRPCClient creates a new client instance connected to the specified address
// establishing a connection to the aggregator RPC server
func NewAggregatorRPCClient(address string, logger log.Logger) (*AggregatorRPCClient, error) {
	manager, err := rpcclientmanager.NewRPCClientManager(
		address,
		logger,
		rpcclientmanager.WithHealthCheckMethod(RPCHealthCheckMethod),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to aggregator: %v", err)
	}

	return &AggregatorRPCClient{
		clientManager: manager,
		logger:        logger,
	}, nil
}

// CollectResponseSignature implements the Aggregator interface by forwarding the request
// to the RPC server and receiving the validated response
func (ra *AggregatorRPCClient) CollectResponseSignature(responseWithSignature *aggtypes.ResponseWithSignature,
	validatedResponseCh chan<- aggtypes.ValidatedResponse) error {
	var result aggtypes.ValidatedResponse
	ctx := context.Background()
	client, err := ra.clientManager.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get RPC client: %v", err)
	}
	ra.logger.Info("AggregatorClient: Calling RPC method to collect responseWithSignature signature",
		"ResponseWithSignature", responseWithSignature,
	)
	err = client.Call(RPCServerAggregatorMethod, responseWithSignature, &result)

	if errors.Is(err, rpc.ErrShutdown) {
		// If the client is shutdown, try to reconnect
		ra.logger.Info("RPC client is shutdown, attempting to reconnect")
		client, err = ra.clientManager.GetClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to get RPC client after shutdown: %v", err)
		}
		err = client.Call(RPCServerAggregatorMethod, responseWithSignature, &result)
	}

	if err != nil {
		return fmt.Errorf("failed to call aggregator RPC method: %v", err)
	}

	if result.Err != nil {
		return fmt.Errorf("aggregator returned error: %v", result.Err)
	}

	ra.logger.Info("AggregatorClient: Received validated responseWithSignature", "result", result)
	validatedResponseCh <- result
	ra.logger.Info("AggregatorClient: Sent validated responseWithSignature to channel")
	return nil
}

// HealthCheck performs a health check on the aggregator service
func (ra *AggregatorRPCClient) HealthCheck() (bool, error) {
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

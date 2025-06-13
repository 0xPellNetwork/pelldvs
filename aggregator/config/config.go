package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/pelldvs-libs/log"
)

const (
	DefaultAggregatorRPCServer     = "0.0.0.0:26653"
	DefaultOperatorResponseTimeout = 5 * time.Second
)

// ChainID represents a unique identifier for a blockchain network
type ChainID = int64

// AggregatorConfig stores configuration settings for the aggregator service
// including network addresses and timeout values
type AggregatorConfig struct {
	AggregatorRPCServer     string `json:"aggregator_rpc_server"`
	OperatorResponseTimeout string `json:"operator_response_timeout"`
}

// ChainConfig stores chain-specific configuration parameters
// including contract addresses and RPC endpoints
type ChainConfig struct {
	ChainID                     ChainID        `json:"-"`
	RPCURL                      string         `json:"rpc_url"`
	OperatorInfoProviderAddress common.Address `json:"operator_info_provider_address"`
	OperatorKeyManagerAddress   common.Address `json:"operator_key_manager_address"`
	CentralSchedulerAddress     common.Address `json:"central_scheduler_address"`
}

// LoadConfig loads and parses aggregator configuration from a JSON file.
// It reads the file at the given path, deserializes the JSON content,
// and returns the resulting configuration object.
func LoadConfig(filePath string, logger log.Logger) (*AggregatorConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config AggregatorConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	cfg := &config
	cfg.Finalize(logger)

	return &config, nil
}

// Finalize ensures that the AggregatorConfig has valid values.
func (c *AggregatorConfig) Finalize(logger log.Logger) {
	if c.AggregatorRPCServer == "" {
		logger.Warn("AggregatorConfig: Aggregator RPC server address is not set, using default",
			"value", DefaultAggregatorRPCServer)
		c.AggregatorRPCServer = DefaultAggregatorRPCServer
	}
	if c.OperatorResponseTimeout == "" {
		logger.Warn("AggregatorConfig: Operator response timeout is not set, using default",
			"value", DefaultOperatorResponseTimeout)
		c.OperatorResponseTimeout = DefaultOperatorResponseTimeout.String()
	}
}

// GetOperatorResponseTimeout converts the string timeout value to a time.Duration.
// This allows the timeout value to be used directly in timing operations,
// parsing the string format into a proper duration object.
func (c *AggregatorConfig) GetOperatorResponseTimeout(logger log.Logger) (time.Duration, error) {
	timeout, err := time.ParseDuration(c.OperatorResponseTimeout)
	if err != nil {
		logger.Error("Invalid operator response timeout", "error", err, "value", c.OperatorResponseTimeout)
		return DefaultOperatorResponseTimeout, fmt.Errorf("invalid operator response timeout: %v", err)
	}
	if timeout < DefaultOperatorResponseTimeout {
		logger.Warn("Operator response timeout is less than default", "value", timeout)
		timeout = DefaultOperatorResponseTimeout
	}
	return timeout, nil
}

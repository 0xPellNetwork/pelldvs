package aggregator

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
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
func LoadConfig(filePath string) (*AggregatorConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config AggregatorConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

// GetOperatorResponseTimeout converts the string timeout value to a time.Duration.
// This allows the timeout value to be used directly in timing operations,
// parsing the string format into a proper duration object.
func (c *AggregatorConfig) GetOperatorResponseTimeout() (time.Duration, error) {
	return time.ParseDuration(c.OperatorResponseTimeout)
}

package aggregator

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// ChainID represents a blockchain network identifier
type ChainID = int64

// AggregatorConfig contains the configuration for the signature aggregation service
// including network settings and timeout parameters
type AggregatorConfig struct {
	AggregatorRPCServer     string `json:"aggregator_rpc_server"`
	OperatorResponseTimeout string `json:"operator_response_timeout"`
}

// ChainConfig struct for storing chain-specific configuration
// including contract addresses and RPC endpoints
type ChainConfig struct {
	ChainID                     ChainID        `json:"-"`
	RPCURL                      string         `json:"rpc_url"`
	OperatorInfoProviderAddress common.Address `json:"operator_info_provider_address"`
	OperatorKeyManagerAddress   common.Address `json:"operator_key_manager_address"`
	CentralSchedulerAddress     common.Address `json:"central_scheduler_address"`
}

// LoadConfig loads the aggregator configuration from the specified file path
// parsing the JSON configuration into an AggregatorConfig struct
func LoadConfig(filePath string) (*AggregatorConfig, error) {
	// Read the configuration file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse the JSON data into configuration structure
	var config AggregatorConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

// GetOperatorResponseTimeout converts the string timeout value
// to a time.Duration type for easier use in timing operations
func (c *AggregatorConfig) GetOperatorResponseTimeout() (time.Duration, error) {
	return time.ParseDuration(c.OperatorResponseTimeout)
}

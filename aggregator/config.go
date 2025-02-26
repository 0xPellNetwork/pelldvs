package aggregator

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type ChainID = int64

type AggregatorConfig struct {
	AggregatorRPCServer     string `json:"aggregator_rpc_server"`
	OperatorResponseTimeout string `json:"operator_response_timeout"`
}

// ChainConfig struct for storing chain-specific configuration
type ChainConfig struct {
	ChainID                     ChainID        `json:"-"`
	RPCURL                      string         `json:"rpc_url"`
	OperatorInfoProviderAddress common.Address `json:"operator_info_provider_address"`
	OperatorKeyManagerAddress   common.Address `json:"operator_key_manager_address"`
	CentralSchedulerAddress     common.Address `json:"central_scheduler_address"`
}

// LoadConfig loads the aggregator configuration from the specified file path
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

// GetOperatorResponseTimeout returns OperatorResponseTimeout as time.Duration
func (c *AggregatorConfig) GetOperatorResponseTimeout() (time.Duration, error) {
	return time.ParseDuration(c.OperatorResponseTimeout)
}

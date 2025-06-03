package rpcclientmanager

import (
	"fmt"
	"net/rpc"
	"sync"
	"time"

	"github.com/0xPellNetwork/pelldvs-libs/log"
)

// RPCClientManager is responsible for managing the RPC client connection to the RPC server.
type RPCClientManager struct {
	address           string
	client            *rpc.Client
	lock              sync.RWMutex
	logger            log.Logger
	lastCheck         time.Time
	checkInterval     time.Duration
	healthCheckMethod string
}

// NewRPCClientManager creates a new RPCClientManager instance.
func NewRPCClientManager(address string, healthCheckMethod string, logger log.Logger) (*RPCClientManager, error) {
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}
	if healthCheckMethod == "" {
		return nil, fmt.Errorf("healthCheckMethod cannot be empty")
	}

	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to aggregator: %v", err)
	}
	return &RPCClientManager{
		address:           address,
		client:            client,
		logger:            logger,
		checkInterval:     500 * time.Millisecond,
		healthCheckMethod: healthCheckMethod,
	}, nil
}

// GetClient returns an existing RPC client if it's still valid, otherwise creates a new one.
func (m *RPCClientManager) GetClient() (*rpc.Client, error) {
	// try to return the existing client if it's still valid
	m.lock.RLock()
	if m.client != nil && time.Since(m.lastCheck) < m.checkInterval {
		client := m.client
		m.lock.RUnlock()
		return client, nil
	}
	m.lock.RUnlock()

	// need to check or create a new client
	m.lock.Lock()
	defer m.lock.Unlock()

	// double-check if the client is still valid
	if m.client != nil && time.Since(m.lastCheck) < m.checkInterval {
		return m.client, nil
	}

	// check if the client is healthy
	if m.client != nil {
		var result bool
		err := m.client.Call(m.healthCheckMethod, struct{}{}, &result)
		if err == nil && result {
			m.lastCheck = time.Now()
			return m.client, nil
		}

		// client unhealthy, close it
		m.logger.Info("RPC client unhealthy, reconnecting",
			"address", m.address, "error", err)
		_ = m.client.Close()
		m.client = nil
	}

	// create a new client
	var err error
	var newClient *rpc.Client

	// try to connect multiple times
	for i := 0; i < 3; i++ {
		newClient, err = rpc.Dial("tcp", m.address)

		if err == nil {
			// check if the new client is healthy
			var result bool
			checkErr := newClient.Call(m.healthCheckMethod, struct{}{}, &result)
			if checkErr == nil && result {
				break
			}

			if checkErr != nil {
				err = checkErr
				_ = newClient.Close()
			} else {
				err = fmt.Errorf("service reports unhealthy")
				_ = newClient.Close()
			}
		}

		m.logger.Error("Failed to connect to RPC server, retrying",
			"address", m.address, "attempt", i+1, "error", err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC server after retries: %w", err)
	}

	m.client = newClient
	m.lastCheck = time.Now()
	m.logger.Info("Successfully connected to RPC server", "address", m.address)

	return m.client, nil
}

func (m *RPCClientManager) SetCheckInterval(interval time.Duration) {
	m.checkInterval = interval
}

// Close closes the RPC client connection
func (m *RPCClientManager) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.client != nil {
		err := m.client.Close()
		if err != nil {
			return err
		}
		m.client = nil
	}
	return nil
}

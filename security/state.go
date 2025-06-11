// Package security provides functionality for managing the security aspects
// of the distributed validation system, including request storage and operator state
package security

import (
	"encoding/json"
	"fmt"

	dbm "github.com/cosmos/cosmos-db"

	"github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/types"
)

// RequestStore defines the interface for storing and retrieving DVS requests
// providing persistence capabilities for the validation system
type RequestStore interface {
	StoreRequest(req *DVSReqResponse) error
	FetchRequest(id string) (*RequestProcessRequest, error)
}

// DVSState maintains the current state of a DVS node,
// including operator identity and request storage
type DVSState struct {
	operatorID   types.OperatorID
	requestStore RequestStore
}

// NewDVSState creates a new DVSState instance initialized with
// the operator's identity and a storage implementation
func NewDVSState(cfg *config.PellConfig, requestStore RequestStore, storeDir string) (*DVSState, error) {
	// Get operator address from the ECDSA private key
	operatorAddress, err := ecdsa.GetAddressFromKeyStoreFile(cfg.OperatorECDSAPrivateKeyStorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator address: %v", err)
	}

	// Generate operator ID from the address
	operatorID := types.OperatorIDFromAddress(operatorAddress)

	// If no requestStore is provided, create a local storage implementation
	if requestStore == nil {
		var err error
		requestStore, err = NewPersistentStore(storeDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create DVS request store: %v", err)
		}
	}

	return &DVSState{
		operatorID:   operatorID,
		requestStore: requestStore,
	}, nil
}

// StoreRequest delegates the request saving operation to the underlying store
func (dvsState *DVSState) StoreRequest(req *DVSReqResponse) error {
	return dvsState.requestStore.StoreRequest(req)
}

// PersistentStore implements RequestStore using a persistent database backend
// for reliable storage of DVS requests and responses
type PersistentStore struct {
	db dbm.DB
}

// NewPersistentStore creates a new PersistentStore instance with a LevelDB backend
// in the specified directory
func NewPersistentStore(dir string) (*PersistentStore, error) {
	db, err := dbm.NewDB("dvs_req_store", dbm.GoLevelDBBackend, dir)
	if err != nil {
		return nil, err
	}

	return &PersistentStore{db: db}, nil
}

// StoreRequest persists a DVS request and its response to the database
// using the request data as the key
func (s *PersistentStore) StoreRequest(req *DVSReqResponse) error {
	key := []byte(fmt.Sprintf("%x", req.Request.Data))
	value, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %v", err)
	}
	err = s.db.Set(key, value)
	if err != nil {
		return fmt.Errorf("failed to save request: %v", err)
	}

	fmt.Printf("Request successfully stored. Key: %x, Value: %s\n", key, string(value))
	fmt.Printf("Request details - Data: %x, Response Hash: %x\n", req.Request.Data, req.Response.Hash)
	return nil
}

// FetchRequest retrieves a previously stored request from the database
// based on the provided identifier
func (s *PersistentStore) FetchRequest(id string) (*RequestProcessRequest, error) {
	key := []byte(id)
	value, err := s.db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get request: %v", err)
	}
	if value == nil {
		return nil, fmt.Errorf("request not found")
	}

	var req RequestProcessRequest
	if err := json.Unmarshal(value, &req); err != nil {
		return nil, fmt.Errorf("failed to deserialize request: %v", err)
	}

	return &req, nil
}

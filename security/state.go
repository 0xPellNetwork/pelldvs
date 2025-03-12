package security

import (
	"encoding/json"
	"fmt"

	dbm "github.com/cometbft/cometbft-db"

	"github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/crypto/ecdsa"
	"github.com/0xPellNetwork/pelldvs/types"
)

type DVSReqStore interface {
	SaveReq(req *DVSReqResponse) error
	GetReq(id string) (*RequestProcessRequest, error)
}

type DVSState struct {
	operatorID  types.OperatorID
	dvsReqStore DVSReqStore
}

// NewDVSState creates a new DVSState instance
func NewDVSState(cfg *config.PellConfig, dvsReqStore DVSReqStore, storeDir string) (*DVSState, error) {
	// get operator address
	operatorAddress, err := ecdsa.GetAddressFromKeyStoreFile(cfg.OperatorECDSAPrivateKeyStorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator address: %v", err)
	}

	// generate operatorID
	operatorID := types.GenOperatorIDByAddress(operatorAddress)

	// If no dvsReqStore is provided, create a local storage implementation
	if dvsReqStore == nil {
		var err error
		dvsReqStore, err = NewStore(storeDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create DVS request store: %v", err)
		}
	}

	return &DVSState{
		operatorID:  operatorID,
		dvsReqStore: dvsReqStore,
	}, nil
}

func (dvsState *DVSState) SaveReq(req *DVSReqResponse) error {
	return dvsState.dvsReqStore.SaveReq(req)
}

// Store represents the persistent storage for DVS requests
type Store struct {
	db dbm.DB
}

// NewStore creates a new Store instance
func NewStore(dir string) (*Store, error) {
	db, err := dbm.NewDB("dvs_req_store", "goleveldb", dir)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) SaveReq(req *DVSReqResponse) error {
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

func (s *Store) GetReq(id string) (*RequestProcessRequest, error) {
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

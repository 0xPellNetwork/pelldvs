package rpc

import (
	"math/big"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"

	interactorcfg "github.com/0xPellNetwork/pelldvs-interactor/config"
	"github.com/0xPellNetwork/pelldvs-interactor/interactor/reader"
	"github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/aggregator"
	"github.com/0xPellNetwork/pelldvs/libs/service"
)

// Task represents an ongoing signature collection and aggregation job.
// It tracks operator responses, manages communication channels, and stores
// state information needed for the aggregation process including operator
// information, group mappings, and threshold requirements.
type Task struct {
	operatorResponses     map[types.OperatorID]aggregator.ResponseWithSignature
	responsesChan         chan aggregator.ResponseWithSignature
	done                  chan aggregator.ValidatedResponse
	timer                 *time.Timer
	taskID                string
	blockNumber           uint32
	chainConfig           *interactorcfg.DVSConfig
	digestToOperators     map[ResultDigest][]types.OperatorID
	operatorStateInfo     *reader.OperatorStateInfo
	operatorsDvsStateDict map[types.OperatorID]types.OperatorDVSState
	groupOperatorMap      map[types.GroupNumber]types.GroupDVSState
	groupNumbers          types.GroupNumbers
	thresholdPercentages  types.GroupThresholdPercentages
}

// RPCServerAggregator implements the Aggregator interface over RPC.
// It manages signature collection tasks, provides thread-safe access to shared
// resources, handles network communication, and coordinates the entire
// aggregation workflow for distributed validation requests.
type RPCServerAggregator struct {
	service.BaseService
	tasks                   map[string]*Task
	tasksMutex              sync.RWMutex
	tasksLocks              map[string]*sync.Mutex
	operatorResponseTimeout time.Duration
	server                  *rpc.Server
	listener                net.Listener
	rpcAddress              string
	chainConfigs            map[uint64]*interactorcfg.DVSConfig
	dvsReader               reader.DVSReader
	Logger                  log.Logger
}

// ResultDigest represents a 32-byte hash of a response result.
// It serves as a unique identifier for operator response consensus
// and is used to group matching responses during aggregation.
type ResultDigest [32]byte

// OperatorStakeInfo holds information about an operator's stake.
// It combines the operator's blockchain address, system identifier,
// and current stake amount for use in threshold calculations.
type OperatorStakeInfo struct {
	Operator   common.Address
	OperatorID types.OperatorID
	Stake      *big.Int
}

// OperatorStateInfo contains the current state of operators in the system.
// It tracks operator mappings, group stake totals, and the distribution
// of operators across validation groups for consensus determination.
type OperatorStateInfo struct {
	Operators        map[types.OperatorID]common.Address
	GroupStakes      map[types.GroupNumber]*big.Int
	GroupOperatorMap map[types.GroupNumber][]OperatorStakeInfo
}

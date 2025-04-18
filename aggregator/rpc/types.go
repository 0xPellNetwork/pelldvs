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
	"github.com/panjf2000/ants/v2"
)

// Task represents a signature collection and aggregation job for a specific DVS request
type Task struct {
	taskID      string      // Unique identifier for this task
	blockNumber uint32      // Block height at which the request was made
	timer       *time.Timer // Timer for auto-finalizing the task after timeout

	// Configuration and state information
	chainConfig          *interactorcfg.DVSConfig        // Chain-specific configuration
	operatorStateInfo    *reader.OperatorStateInfo       // Information about operators active at this block
	groupNumbers         types.GroupNumbers              // Groups involved in this request
	thresholdPercentages types.GroupThresholdPercentages // Required signature thresholds for each group

	// Data collected during task processing
	operatorResponses     map[types.OperatorID]aggregator.ResponseWithSignature // Responses from each operator
	digestToOperators     map[ResultDigest][]types.OperatorID                   // Mapping of digest to operators who signed it
	operatorsDvsStateDict map[types.OperatorID]types.OperatorDVSState           // DVS state for each operator
	groupOperatorMap      map[types.GroupNumber]types.GroupDVSState             // DVS state for each group

	// Communication channels
	responsesChan chan aggregator.ResponseWithSignature // Channel for collecting incoming operator responses
	done          chan aggregator.ValidatedResponse     // Channel for sending aggregation results back to operators
}

// RPCServerAggregator is an RPC implementation of the Aggregator interface
// responsible for collecting, validating and aggregating operator signatures
type RPCServerAggregator struct {
	service.BaseService
	tasks                   sync.Map                            // Maps task IDs to active Task instances
	tasksLocks              sync.Map                            // Maps task IDs to mutex locks for thread safety
	operatorResponseTimeout time.Duration                       // How long to wait for operator responses
	server                  *rpc.Server                         // RPC server for handling operator requests
	listener                net.Listener                        // Network listener for incoming connections
	rpcAddress              string                              // Address where the RPC server listens
	chainConfigs            map[uint64]*interactorcfg.DVSConfig // Chain-specific configurations
	dvsReader               reader.DVSReader                    // Interface for reading DVS state from blockchain
	Logger                  log.Logger                          // Logger for recording events and errors
	taskPool                *ants.Pool                          // Worker pool for efficient task processing
}

// ResultDigest is a 32-byte hash representing the digest of a response
// Used to identify and group consensus among operator responses
type ResultDigest [32]byte

// OperatorStakeInfo contains information about an operator's stake
// in the distributed validation system
type OperatorStakeInfo struct {
	Operator   common.Address   // Ethereum address of the operator
	OperatorID types.OperatorID // Unique identifier for the operator
	Stake      *big.Int         // Stake amount of this operator
}

// OperatorStateInfo encapsulates the state of all operators
// in the system at a specific point in time
type OperatorStateInfo struct {
	Operators        map[types.OperatorID]common.Address       // Maps operator IDs to their addresses
	GroupStakes      map[types.GroupNumber]*big.Int            // Total stake for each group
	GroupOperatorMap map[types.GroupNumber][]OperatorStakeInfo // Operators in each group with their stakes
}

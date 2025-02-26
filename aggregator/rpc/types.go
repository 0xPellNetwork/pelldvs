package rpc

import (
	"math/big"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"

	interactorcfg "github.com/0xPellNetwork/pelldvs-interactor/config"
	"github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/aggregator"
	"github.com/0xPellNetwork/pelldvs/libs/service"
)

type Task struct {
	operatorResponses     map[types.OperatorID]aggregator.ResponseWithSignature
	responsesChan         chan aggregator.ResponseWithSignature
	done                  chan aggregator.ValidatedResponse
	timer                 *time.Timer
	taskID                string
	blockNumber           uint32
	chainConfig           *interactorcfg.DVSConfig
	digestToOperators     map[ResultDigest][]types.OperatorID
	operatorStateInfo     *OperatorStateInfo
	operatorsDvsStateDict map[types.OperatorID]types.OperatorDVSState
	groupOperatorMap      map[types.GroupNumber]types.GroupDVSState
	groupNumbers          types.GroupNumbers
	thresholdPercentages  types.GroupThresholdPercentages
}

// RPCServerAggregator is an RPC implementation of the Aggregator interface
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
	dvsInteractor           map[uint64]*AggregatorDVSInteractor
	Logger                  log.Logger
}

type ResultDigest [32]byte

type OperatorStakeInfo struct {
	Operator   common.Address
	OperatorID types.OperatorID
	Stake      *big.Int
}

type OperatorStateInfo struct {
	Operators        map[types.OperatorID]common.Address
	GroupStakes      map[types.GroupNumber]*big.Int
	GroupOperatorMap map[types.GroupNumber][]OperatorStakeInfo
}

package rpc

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/rpc"
	"sort"
	"sync"
	"time"

	dbm "github.com/cosmos/cosmos-db"

	interactorcfg "github.com/0xPellNetwork/pelldvs-interactor/config"
	"github.com/0xPellNetwork/pelldvs-interactor/interactor/reader"
	"github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs-libs/crypto/bls"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	aggcfg "github.com/0xPellNetwork/pelldvs/aggregator/config"
	aggtypes "github.com/0xPellNetwork/pelldvs/aggregator/types"
	avsitypes "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/libs/service"
	"github.com/0xPellNetwork/pelldvs/rpc/core/errcode"
	rpctypes "github.com/0xPellNetwork/pelldvs/rpc/jsonrpc/types"
)

// DBContext holds the necessary information for initializing a database
// including configuration and identifier
type DBContext struct {
	ID     string
	Config *config.Config
}

// DBProvider defines a function type that creates and returns a database
// based on the provided context
type DBProvider func(*DBContext) (dbm.DB, error)

// DefaultDBProvider creates a database using the configuration specified in the context
// implementing the standard database initialization process
func DefaultDBProvider(ctx *DBContext) (dbm.DB, error) {
	dbType := dbm.BackendType(ctx.Config.DBBackend)
	return dbm.NewDB(ctx.ID, dbType, ctx.Config.DBDir())
}

// NewAggregatorGRPCServer creates a new instance of the RPC server aggregator
// initializing all required components and connections
func NewAggregatorGRPCServer(
	ctx context.Context,
	aggConfig *aggcfg.AggregatorConfig,
	interactorConfig *interactorcfg.Config,
	cfg *config.Config,
	dvsReader reader.DVSReader,
	logger log.Logger,
) (*AggregatorRPCServer, error) {
	timeout, _ := aggConfig.GetOperatorResponseTimeout()
	tasksLocks := make(map[string]*sync.Mutex)
	ra := &AggregatorRPCServer{
		tasks:                   make(map[string]*Task),
		operatorResponseTimeout: timeout,
		server:                  rpc.NewServer(),
		rpcAddress:              aggConfig.AggregatorRPCServer,
		chainConfigs:            interactorConfig.ContractConfig.DVSConfigs,
		tasksLocks:              tasksLocks,
		logger:                  logger.With("module", "RPCServerAggregatorServer"),
		dvsReader:               dvsReader,
	}
	ra.BaseService = *service.NewBaseService(nil, "AggregatorRPCServer", ra)

	ra.logger.Info("NewAggregatorGRPCServer initialized", "timeout", ra.operatorResponseTimeout)

	return ra, nil
}

// OnStart initializes and starts the RPC server
// registering handlers and beginning to accept connections
func (ra *AggregatorRPCServer) OnStart() error {
	if err := ra.server.Register(ra); err != nil {
		return fmt.Errorf("failed to register RPC handler: %v", err)
	}

	listener, err := net.Listen("tcp", ra.rpcAddress)
	if err != nil {
		return fmt.Errorf("unable to listen on address %s: %v", ra.rpcAddress, err)
	}

	ra.listener = listener
	ra.logger.Info("RPC server started", "address", ra.rpcAddress)

	go ra.server.Accept(listener)
	return nil
}

// IsRunning checks if the server is currently running
// implementing the service interface requirement
func (ra *AggregatorRPCServer) IsRunning() bool {
	return ra.BaseService.IsRunning()
}

// HealthCheck provides a simple health check for the RPC server
func (ra *AggregatorRPCServer) HealthCheck(_ struct{}, reply *bool) error {
	*reply = ra.IsRunning()
	return nil
}

// OnStop gracefully shuts down the RPC server
// closing the network listener
func (ra *AggregatorRPCServer) OnStop() {
	if ra.listener != nil {
		ra.listener.Close()
	}
}

// CollectResponseSignature processes operator signature submissions
// creating or updating tasks and managing the aggregation process
func (ra *AggregatorRPCServer) CollectResponseSignature(response *aggtypes.ResponseWithSignature,
	result *aggtypes.ValidatedResponse) error {
	taskID := ra.generateTaskID(response.RequestData)
	ra.logger.Info("CollectResponseSignature start",
		"taskID", taskID, "operatorID", response.OperatorID,
		"response", response,
		"result", result,
	)

	ra.tasksMutex.Lock()
	taskLock, exists := ra.tasksLocks[taskID]
	if !exists {
		taskLock = &sync.Mutex{}
		ra.tasksLocks[taskID] = taskLock
	}
	ra.tasksMutex.Unlock()

	taskLock.Lock()
	defer taskLock.Unlock()

	task, exists := ra.tasks[taskID]
	if !exists {
		chainID := big.NewInt(response.RequestData.ChainId)
		chainConfig, ok := ra.chainConfigs[chainID.Uint64()]
		if !ok {
			return fmt.Errorf("chain config not found for chain ID: %s", chainID.String())
		}

		groupNumbers := types.GroupNumbers{}
		for _, groupNumber := range response.RequestData.GroupNumbers {
			groupNumbers = append(groupNumbers, types.GroupNumber(groupNumber))
		}

		thresholdPercentages := types.GroupThresholdPercentages{}
		for _, thresholdPercentage := range response.RequestData.GroupThresholdPercentages {
			thresholdPercentages = append(thresholdPercentages, types.GroupThresholdPercentage(thresholdPercentage))
		}

		blockNumber := uint32(response.RequestData.Height)

		operatorsDvsStateDict, err := ra.dvsReader.GetOperatorsDVSStateAtBlock(chainID.Uint64(), groupNumbers, blockNumber)
		if err != nil {
			ra.logger.Error("Failed to get operators DVS state", "block", blockNumber, "error", err)
			return err
		}

		groupsDvsStateDict, err := ra.dvsReader.GetGroupsDVSStateAtBlock(chainID.Uint64(), groupNumbers, blockNumber)
		if err != nil {
			ra.logger.Error("Failed to get groups DVS state", "block", blockNumber, "error", err)
			return err
		}

		operatorStateInfo, err := ra.dvsReader.GetOperatorState(chainID.Uint64(), groupNumbers, blockNumber)
		if err != nil {
			ra.logger.Error("Failed to get operator state", "error", err)
			return fmt.Errorf("failed to get operator state: %v", err)
		}

		task = &Task{
			operatorResponses:     make(map[types.OperatorID]aggtypes.ResponseWithSignature),
			responsesChan:         make(chan aggtypes.ResponseWithSignature, len(operatorStateInfo.Operators)),
			done:                  make(chan aggtypes.ValidatedResponse, len(operatorStateInfo.Operators)),
			taskID:                taskID,
			chainConfig:           chainConfig,
			digestToOperators:     make(map[ResultDigest][]types.OperatorID),
			operatorStateInfo:     operatorStateInfo,
			groupOperatorMap:      groupsDvsStateDict,
			operatorsDvsStateDict: operatorsDvsStateDict,
			groupNumbers:          groupNumbers,
			thresholdPercentages:  thresholdPercentages,
			blockNumber:           blockNumber,
		}
		ra.tasks[taskID] = task

		ra.logger.Info("New task created",
			"taskID", taskID,
			"chainID", chainID,
			"blockNumber", blockNumber,
			"groupNumbers", groupNumbers,
			"thresholdPercentages", thresholdPercentages,
			"operatorsCount", len(operatorStateInfo.Operators),
			"operatorsDvsStateDict", operatorsDvsStateDict,
			"groupsDvsStateDict", groupsDvsStateDict,
			"operatorStateInfo", operatorStateInfo,
		)
		ra.logger.Info("Task created and we will finalize it after timeout",
			"taskID", taskID,
			"operatorResponseTimeout", ra.operatorResponseTimeout,
		)
		task.timer = time.AfterFunc(ra.operatorResponseTimeout, func() {
			ra.logger.Info("Timer triggered, calling finalizeTask", "taskID", taskID, "timeout", ra.operatorResponseTimeout)
			ra.finalizeTask(taskID)
		})
		go ra.processResponses(task)
	} else {
		ra.logger.Info("Task already exists", "taskID", taskID)
	}

	ra.logger.Info("Adding response to the shared channel",
		"taskID", taskID, "operatorID", response.OperatorID)
	task.responsesChan <- *response

	ra.logger.Info("Waiting for task result",
		"taskID", taskID, "operatorID", response.OperatorID)
	validatedResponse := <-task.done

	*result = validatedResponse

	ra.logger.Info("CollectResponseSignature done",
		"taskID", taskID,
		"operatorID", response.OperatorID,
		"result", result,
	)

	return nil
}

func (ra *AggregatorRPCServer) generateTaskID(request avsitypes.DVSRequest) string {
	return hex.EncodeToString(request.Hash())
}

func (ra *AggregatorRPCServer) processResponses(task *Task) {
	ra.logger.Info("processResponses started", "taskID", task.taskID)
	for response := range task.responsesChan {
		ra.logger.Info("processResponses. Processing response",
			"taskID", task.taskID, "operatorID", response.OperatorID)

		_, err := ra.dvsReader.GetOperatorInfoByID(response.OperatorID)
		if err != nil {
			ra.logger.Error("Failed to get operator info",
				"taskID", task.taskID, "operatorID", response.OperatorID,
				"error", err,
			)
			continue
		}

		task.operatorResponses[response.OperatorID] = response
		task.digestToOperators[response.Digest] = append(task.digestToOperators[response.Digest], response.OperatorID)
		time.Sleep(1 * time.Second)
	}
	ra.logger.Info("processResponses.done", "taskID", task.taskID)
}

func (ra *AggregatorRPCServer) finalizeTask(taskID string) {
	ra.logger.Info("finalizeTask started", "taskID", taskID)
	ra.tasksMutex.Lock()
	defer ra.tasksMutex.Unlock()

	task, exists := ra.tasks[taskID]
	if !exists {
		return
	}

	task.timer.Stop()

	aggregatedResult, err := ra.aggregateSignatures(task)
	if err != nil {
		// Log with more context including the task ID
		ra.logger.Error("Failed to aggregate signatures", "taskID", taskID, "error", err)

		// Create a more detailed error response with the original error message
		aggregatedResult = ra.createErrorValidatedResponse(taskID, &rpctypes.RPCError{
			Code:    errcode.AggregationFailed, // Use a more specific error code
			Message: fmt.Sprintf("Failed to aggregate signatures: %v", err),
			Data:    taskID, // Include task ID as context in the error data
		})
	}

	for operatorID := range task.operatorResponses {
		task.done <- *aggregatedResult
		ra.logger.Info("Task finalizeTask task.done for", "taskID", taskID, "operatorID", operatorID)
	}

	close(task.done)
	close(task.responsesChan)

	delete(ra.tasks, taskID)

	ra.logger.Info("Task deleted", "taskID", taskID)
}

func (ra *AggregatorRPCServer) createErrorValidatedResponse(taskID string,
	err *rpctypes.RPCError) *aggtypes.ValidatedResponse {
	return &aggtypes.ValidatedResponse{
		Data:                        []byte{},
		Err:                         err,
		Hash:                        []byte(taskID),
		NonSignersPubkeysG1:         []*bls.G1Point{},
		GroupApksG1:                 []*bls.G1Point{},
		SignersApkG2:                bls.NewZeroG2Point(),
		SignersAggSigG1:             bls.NewZeroSignature(),
		NonSignerGroupBitmapIndices: []uint32{},
		GroupApkIndices:             []uint32{},
		TotalStakeIndices:           []uint32{},
		NonSignerStakeIndices:       [][]uint32{},
	}
}

func (ra *AggregatorRPCServer) aggregateSignatures(task *Task) (*aggtypes.ValidatedResponse, error) {
	ra.logger.Info("aggregateSignatures.start", "taskID", task.taskID)
	if len(task.operatorResponses) == 0 {
		return nil, errors.New("no signatures to aggregate")
	}

	totalStakePerGroup := make(map[types.GroupNumber]*big.Int)
	for groupNum, groupDvsState := range task.groupOperatorMap {
		totalStakePerGroup[groupNum] = groupDvsState.TotalStake
	}
	var selectedDigest ResultDigest
	var selectedData []byte

	thresholdPercentagesMap := make(map[types.GroupNumber]types.GroupThresholdPercentage)
	for groupNum, thresholdPercentage := range task.thresholdPercentages {
		thresholdPercentagesMap[types.GroupNumber(groupNum)] = thresholdPercentage
	}

	signersTotalStakePerGroup := make(map[types.GroupNumber]*big.Int)

	for digest, operators := range task.digestToOperators {
		for _, addrOperatorID := range operators {
			for groupNumber, stake := range task.operatorsDvsStateDict[addrOperatorID].StakePerGroup {
				if _, ok := signersTotalStakePerGroup[groupNumber]; !ok {
					signersTotalStakePerGroup[groupNumber] = big.NewInt(0)
				}
				signersTotalStakePerGroup[groupNumber].Add(signersTotalStakePerGroup[groupNumber], stake)
			}
			if ra.checkIfStakeThresholdsMet(signersTotalStakePerGroup, totalStakePerGroup, thresholdPercentagesMap) {
				selectedDigest = digest
				selectedData = task.operatorResponses[operators[0]].Data
				ra.logger.Debug("Selected digest for aggregation checkIfStakeThresholdsMet", "digest", selectedDigest)
				break
			} else {
				ra.logger.Error("stake thresholds not met for digest", "digest", digest)
				return nil, fmt.Errorf("stake thresholds not met for digest: %v", digest)
			}
		}
	}

	ra.logger.Debug("Selected digest for aggregation", "digest", selectedDigest)

	aggregatedSignature := bls.NewZeroSignature()
	signersApkG2 := bls.NewZeroG2Point()
	nonSignersOperatorIds := []types.OperatorID{}

	groupApksG1 := make([]*bls.G1Point, 0)
	for _, groupNumber := range task.groupNumbers {
		groupApksG1 = append(groupApksG1, task.groupOperatorMap[groupNumber].AggPubkeyG1)
	}

	for _, response := range task.operatorResponses {
		if response.Digest == selectedDigest {
			addrOperatorID := response.OperatorID
			operator, _ := ra.dvsReader.GetOperatorInfoByID(addrOperatorID)
			aggregatedSignature.Add(response.Signature)
			signersApkG2.Add(operator.Pubkeys.G2Pubkey)
		}
	}

	activeGroups := make(map[types.GroupNumber]bool)

	// Registered operators for groups and block height in a task context
	registeredOperators := make(map[types.OperatorID]types.OperatorInfo)
	operatorMapByBLSOperatorID := make(map[types.OperatorID]types.OperatorInfo)
	for groupNum, operatorInfos := range task.operatorStateInfo.GroupOperatorMap {
		if len(operatorInfos) > 0 {
			activeGroups[groupNum] = true
		}
		for _, operator := range operatorInfos {
			addrOperatorID := operator.OperatorID
			operatorInfo, err := ra.dvsReader.GetOperatorInfoByID(addrOperatorID)
			if err != nil {
				return nil, fmt.Errorf("failed to get operator info by ID: %v", err)
			}
			registeredOperators[addrOperatorID] = operatorInfo

			blsOperatorID := operatorInfo.Pubkeys.GetOperatorID()
			operatorMapByBLSOperatorID[blsOperatorID] = operatorInfo
		}
	}

	ra.logger.Debug("All registered operators in groups",
		"operatorCount", len(registeredOperators),
		"operators", registeredOperators,
	)

	for addrOperatorID := range registeredOperators {
		ra.logger.Debug("Checking operator", "operatorID", addrOperatorID)

		if _, signed := task.operatorResponses[addrOperatorID]; !signed ||
			task.operatorResponses[addrOperatorID].Digest != selectedDigest {
			isInActiveGroup := false
			for groupNum := range task.groupNumbers {
				if activeGroups[types.GroupNumber(groupNum)] {
					isInActiveGroup = true
					break
				}
			}
			if isInActiveGroup {
				blsOperatorID := registeredOperators[addrOperatorID].Pubkeys.GetOperatorID()
				nonSignersOperatorIds = append(nonSignersOperatorIds, blsOperatorID)
			}
		}
	}

	sort.SliceStable(nonSignersOperatorIds, func(i, j int) bool {
		iOprInt := new(big.Int).SetBytes(nonSignersOperatorIds[i][:])
		jOprInt := new(big.Int).SetBytes(nonSignersOperatorIds[j][:])
		return iOprInt.Cmp(jOprInt) == -1
	})

	nonSignersPubkeysG1 := []*bls.G1Point{}
	for _, blsOperatorID := range nonSignersOperatorIds {
		operator, ok := operatorMapByBLSOperatorID[blsOperatorID]
		if !ok {
			ra.logger.Error("Operator not found by blsOperatorID", "blsOperatorID", blsOperatorID)
			continue
		}
		nonSignersPubkeysG1 = append(nonSignersPubkeysG1, operator.Pubkeys.G1Pubkey)
	}

	indices, err := ra.dvsReader.GetCheckSignaturesIndices(
		task.chainConfig.ChainID,
		task.blockNumber,
		task.groupNumbers,
		nonSignersOperatorIds,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get check signatures indices: %v", err)
	}
	ra.logger.Debug("aggregateSignatures.indices", "indices", indices)

	ra.logger.Info("Task aggregation completed successfully", "taskHash", task.taskID, "data", selectedData)
	ra.logger.Debug("Aggregated response details",
		"nonSignersPubkeysG1Count", len(nonSignersPubkeysG1),
		"groupApksG1Count", len(groupApksG1),
		"signersApkG2", signersApkG2,
		"signersAggSigG1", aggregatedSignature,
		"nonSignerGroupBitmapIndices", indices.NonSignerGroupBitmapIndices,
		"groupApkIndices", indices.GroupApkIndices,
		"totalStakeIndices", indices.TotalStakeIndices,
		"nonSignerStakeIndicesCount", len(indices.NonSignerStakeIndices),
	)
	result := &aggtypes.ValidatedResponse{
		Err:                         nil,
		Hash:                        []byte(task.taskID),
		NonSignersPubkeysG1:         nonSignersPubkeysG1,
		GroupApksG1:                 groupApksG1,
		SignersApkG2:                signersApkG2,
		SignersAggSigG1:             aggregatedSignature,
		NonSignerGroupBitmapIndices: indices.NonSignerGroupBitmapIndices,
		GroupApkIndices:             indices.GroupApkIndices,
		TotalStakeIndices:           indices.TotalStakeIndices,
		NonSignerStakeIndices:       indices.NonSignerStakeIndices,
		Data:                        selectedData,
	}

	ra.logger.Info("aggregateSignatures.result", "result", result)

	return result, nil
}

func (ra *AggregatorRPCServer) checkIfStakeThresholdsMet(
	signedStakePerGroup map[types.GroupNumber]*big.Int,
	totalStakePerGroup map[types.GroupNumber]*big.Int,
	groupThresholdPercentagesMap map[types.GroupNumber]types.GroupThresholdPercentage,
) bool {
	for groupNum, groupThresholdPercentage := range groupThresholdPercentagesMap {
		signedStakeByGroup, ok := signedStakePerGroup[groupNum]
		if !ok {
			return false
		}

		totalStakeByGroup, ok := totalStakePerGroup[groupNum]
		if !ok {
			return false
		}

		signedStake := new(big.Int).Mul(signedStakeByGroup, big.NewInt(100))
		thresholdStake := new(big.Int).Mul(totalStakeByGroup, big.NewInt(int64(groupThresholdPercentage)))

		if signedStake.Cmp(thresholdStake) < 0 {
			return false
		}
	}
	return true
}

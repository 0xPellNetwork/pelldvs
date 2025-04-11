package rpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/rpc"
	"sort"
	"sync"
	"time"

	dbm "github.com/cometbft/cometbft-db"

	interactorcfg "github.com/0xPellNetwork/pelldvs-interactor/config"
	"github.com/0xPellNetwork/pelldvs-interactor/interactor/reader"
	"github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs-libs/crypto/bls"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/aggregator"
	"github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/libs/service"
	rpctypes "github.com/0xPellNetwork/pelldvs/rpc/jsonrpc/types"
)

type DBContext struct {
	ID     string
	Config *config.Config
}

// DBProvider takes a DBContext and returns an instantiated DB.
type DBProvider func(*DBContext) (dbm.DB, error)

// DefaultDBProvider returns a database using the DBBackend and DBDir
// specified in the ctx.Config.
func DefaultDBProvider(ctx *DBContext) (dbm.DB, error) {
	dbType := dbm.BackendType(ctx.Config.DBBackend)
	return dbm.NewDB(ctx.ID, dbType, ctx.Config.DBDir())
}

func NewRPCServerAggregator(
	ctx context.Context,
	cfg *config.Config,
	aggConfig *aggregator.AggregatorConfig,
	logger log.Logger,
) (*RPCServerAggregator, error) {
	db, err := DefaultDBProvider(&DBContext{
		ID:     "indexer",
		Config: cfg,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init db: %v", err)
	}

	interactorConfig, err := interactorcfg.LoadConfig(cfg.Pell.InteractorConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create interactor config from file: %v", err)
	}

	dvsReader, err := reader.NewDVSReaderFromConfig(interactorConfig, db, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create DVS reader: %v", err)
	}

	timeout, err := aggConfig.GetOperatorResponseTimeout()
	if err != nil {
		return nil, fmt.Errorf("failed to get operator response timeout: %v", err)
	}
	tasksLocks := make(map[string]*sync.Mutex)
	ra := &RPCServerAggregator{
		tasks:                   make(map[string]*Task),
		operatorResponseTimeout: timeout,
		server:                  rpc.NewServer(),
		rpcAddress:              aggConfig.AggregatorRPCServer,
		chainConfigs:            interactorConfig.ContractConfig.DVSConfigs,
		tasksLocks:              tasksLocks,
		Logger:                  logger.With("module", "RPCServerAggregatorServer"),
		dvsReader:               dvsReader,
	}
	ra.BaseService = *service.NewBaseService(nil, "RPCServerAggregator", ra)
	return ra, nil
}

func (ra *RPCServerAggregator) OnStart() error {
	if err := ra.server.Register(ra); err != nil {
		return fmt.Errorf("failed to register RPC handler: %v", err)
	}

	listener, err := net.Listen("tcp", ra.rpcAddress)
	if err != nil {
		return fmt.Errorf("unable to listen on address %s: %v", ra.rpcAddress, err)
	}

	ra.listener = listener
	ra.Logger.Info("RPC server started", "address", ra.rpcAddress)

	go ra.server.Accept(listener)
	return nil
}

func (ra *RPCServerAggregator) IsRunning() bool {
	return ra.BaseService.IsRunning()
}

func (ra *RPCServerAggregator) OnStop() {
	if ra.listener != nil {
		ra.listener.Close()
	}
}

func (ra *RPCServerAggregator) CollectResponseSignature(response *aggregator.ResponseWithSignature, result *aggregator.ValidatedResponse) error {
	taskID := ra.generateTaskID(response.RequestData)
	ra.Logger.Info("CollectResponseSignature start",
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
			ra.Logger.Error("Failed to get operators DVS state", "block", blockNumber, "error", err)
			return err
		}

		groupsDvsStateDict, err := ra.dvsReader.GetGroupsDVSStateAtBlock(chainID.Uint64(), groupNumbers, blockNumber)
		if err != nil {
			ra.Logger.Error("Failed to get groups DVS state", "block", blockNumber, "error", err)
			return err
		}

		operatorStateInfo, err := ra.dvsReader.GetOperatorState(chainID.Uint64(), groupNumbers, blockNumber)
		if err != nil {
			ra.Logger.Error("Failed to get operator state", "error", err)
			return fmt.Errorf("failed to get operator state: %v", err)
		}

		task = &Task{
			operatorResponses:     make(map[types.OperatorID]aggregator.ResponseWithSignature),
			responsesChan:         make(chan aggregator.ResponseWithSignature, len(operatorStateInfo.Operators)),
			done:                  make(chan aggregator.ValidatedResponse, len(operatorStateInfo.Operators)),
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
		task.timer = time.AfterFunc(ra.operatorResponseTimeout, func() {
			ra.finalizeTask(taskID)
		})
		go ra.processResponses(task)
	} else {
		ra.Logger.Info("Task already exists", "taskID", taskID)
	}

	task.responsesChan <- *response
	validatedResponse := <-task.done
	*result = validatedResponse

	ra.Logger.Info("CollectResponseSignature done",
		"taskID", taskID,
		"operatorID", response.OperatorID,
		"result", result,
	)

	return nil
}

func (ra *RPCServerAggregator) generateTaskID(request interface{}) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%v", request)))
	return hex.EncodeToString(hash[:])
}

func (ra *RPCServerAggregator) processResponses(task *Task) {
	for response := range task.responsesChan {
		_, err := ra.dvsReader.GetOperatorInfoByID(response.OperatorID)
		if err != nil {
			ra.Logger.Error("Failed to get operator info",
				"taskID", task.taskID, "operatorID", response.OperatorID,
				"error", err,
			)
			continue
		}

		task.operatorResponses[response.OperatorID] = response
		task.digestToOperators[response.Digest] = append(task.digestToOperators[response.Digest], response.OperatorID)
	}
	ra.Logger.Info("processResponses.done", "taskID", task.taskID)
}

func (ra *RPCServerAggregator) finalizeTask(taskID string) {
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
		ra.Logger.Error("Failed to aggregate signatures",
			"taskID", taskID,
			"error", err)

		// Create a more detailed error response with the original error message
		aggregatedResult = ra.createErrorValidatedResponse(taskID, &rpctypes.RPCError{
			Code:    -32000, // Use a more specific error code
			Message: fmt.Sprintf("Failed to aggregate signatures: %v", err),
			Data:    taskID, // Include task ID as context in the error data
		})
	}

	for operatorID := range task.operatorResponses {
		task.done <- *aggregatedResult
		ra.Logger.Info("Task finalizeTask task.done for", "taskID", taskID, "operatorID", operatorID)
	}

	close(task.done)
	close(task.responsesChan)

	delete(ra.tasks, taskID)

	ra.Logger.Info("Task deleted", "taskID", taskID)
}

func (ra *RPCServerAggregator) createErrorValidatedResponse(taskID string, err *rpctypes.RPCError) *aggregator.ValidatedResponse {
	return &aggregator.ValidatedResponse{
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

func (ra *RPCServerAggregator) aggregateSignatures(task *Task) (*aggregator.ValidatedResponse, error) {
	ra.Logger.Info("aggregateSignatures.start", "taskID", task.taskID)
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
				ra.Logger.Debug("Selected digest for aggregation checkIfStakeThresholdsMet", "digest", selectedDigest)
				break
			} else {
				ra.Logger.Error("stake thresholds not met for digest", "digest", digest)
				return nil, fmt.Errorf("stake thresholds not met for digest: %v", digest)
			}
		}
	}

	ra.Logger.Debug("Selected digest for aggregation", "digest", selectedDigest)

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

	ra.Logger.Debug("All registered operators in groups",
		"operatorCount", len(registeredOperators),
		"operators", registeredOperators)

	for addrOperatorID := range registeredOperators {
		if _, signed := task.operatorResponses[addrOperatorID]; !signed || task.operatorResponses[addrOperatorID].Digest != selectedDigest {
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
			ra.Logger.Error("Operator not found by blsOperatorID", "blsOperatorID", blsOperatorID)
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
	ra.Logger.Debug("aggregateSignatures.indices", "indices", indices)

	ra.Logger.Info("Task aggregation completed successfully", "taskHash", task.taskID, "data", selectedData)
	ra.Logger.Debug("Aggregated response details",
		"nonSignersPubkeysG1Count", len(nonSignersPubkeysG1),
		"groupApksG1Count", len(groupApksG1),
		"signersApkG2", signersApkG2,
		"signersAggSigG1", aggregatedSignature,
		"nonSignerGroupBitmapIndices", indices.NonSignerGroupBitmapIndices,
		"groupApkIndices", indices.GroupApkIndices,
		"totalStakeIndices", indices.TotalStakeIndices,
		"nonSignerStakeIndicesCount", len(indices.NonSignerStakeIndices),
	)
	result := &aggregator.ValidatedResponse{
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

	ra.Logger.Info("aggregateSignatures.result", "result", result)

	return result, nil
}

func (ra *RPCServerAggregator) checkIfStakeThresholdsMet(
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

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

	dbm "github.com/cometbft/cometbft-db"
	"github.com/panjf2000/ants/v2"

	interactorcfg "github.com/0xPellNetwork/pelldvs-interactor/config"
	"github.com/0xPellNetwork/pelldvs-interactor/interactor/reader"
	"github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs-libs/crypto/bls"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/aggregator"
	avsitypes "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/libs/service"
	"github.com/0xPellNetwork/pelldvs/rpc/core/errcode"
	rpctypes "github.com/0xPellNetwork/pelldvs/rpc/jsonrpc/types"
)

// Maximum number of concurrent tasks that can be processed in the worker pool
const TaskPoolSize = 128

// DBContext contains database configuration information
type DBContext struct {
	ID     string
	Config *config.Config
}

// DBProvider creates and returns a database instance based on the provided context
type DBProvider func(*DBContext) (dbm.DB, error)

// DefaultDBProvider creates a database using configuration from the context
func DefaultDBProvider(ctx *DBContext) (dbm.DB, error) {
	dbType := dbm.BackendType(ctx.Config.DBBackend)
	return dbm.NewDB(ctx.ID, dbType, ctx.Config.DBDir())
}

// NewRPCServerAggregator initializes and returns a new aggregator server
// that collects signatures from operators and processes them
func NewRPCServerAggregator(
	ctx context.Context,
	cfg *config.Config,
	aggConfig *aggregator.AggregatorConfig,
	logger log.Logger,
) (*RPCServerAggregator, error) {
	// Initialize database for request indexing
	db, err := DefaultDBProvider(&DBContext{
		ID:     "indexer",
		Config: cfg,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init db: %v", err)
	}

	// Load configuration for interacting with blockchain
	interactorConfig, err := interactorcfg.LoadConfig(cfg.Pell.InteractorConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create interactor config from file: %v", err)
	}

	// Set up the reader for retrieving DVS state information
	dvsReader, err := reader.NewDVSReaderFromConfig(interactorConfig, db, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create DVS reader: %v", err)
	}

	// Get configured timeout for operator responses
	timeout, err := aggConfig.GetOperatorResponseTimeout()
	if err != nil {
		return nil, fmt.Errorf("failed to get operator response timeout: %v", err)
	}

	// Create worker pool for handling tasks efficiently
	workerPool, err := ants.NewPool(TaskPoolSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create worker pool: %v", err)
	}

	// Initialize the aggregator server with all necessary components
	ra := &RPCServerAggregator{
		tasks:                   sync.Map{},
		operatorResponseTimeout: timeout,
		server:                  rpc.NewServer(),
		rpcAddress:              aggConfig.AggregatorRPCServer,
		chainConfigs:            interactorConfig.ContractConfig.DVSConfigs,
		tasksLocks:              sync.Map{},
		Logger:                  logger.With("module", "RPCServerAggregatorServer"),
		dvsReader:               dvsReader,
		taskPool:                workerPool,
	}
	ra.BaseService = *service.NewBaseService(nil, "RPCServerAggregator", ra)
	return ra, nil
}

// OnStart registers the RPC server and starts listening for connections
func (ra *RPCServerAggregator) OnStart() error {
	// Register this object to handle RPC calls
	if err := ra.server.Register(ra); err != nil {
		return fmt.Errorf("failed to register RPC handler: %v", err)
	}

	// Start listening for incoming connections
	listener, err := net.Listen("tcp", ra.rpcAddress)
	if err != nil {
		return fmt.Errorf("unable to listen on address %s: %v", ra.rpcAddress, err)
	}

	ra.listener = listener
	ra.Logger.Info("RPC server started", "address", ra.rpcAddress)

	// Begin accepting connections in a separate goroutine
	go ra.server.Accept(listener)
	return nil
}

// IsRunning checks if the service is currently running
func (ra *RPCServerAggregator) IsRunning() bool {
	return ra.BaseService.IsRunning()
}

// OnStop closes the listener when the service is stopped
func (ra *RPCServerAggregator) OnStop() {
	if ra.listener != nil {
		ra.listener.Close()
	}
}

// CollectResponseSignature handles signature collection from operators.
// It creates or retrieves a task for the given request, adds the operator's response,
// and waits for the aggregated result.
func (ra *RPCServerAggregator) CollectResponseSignature(response *aggregator.ResponseWithSignature, result *aggregator.ValidatedResponse) error {
	// Generate a unique task ID based on the request hash
	taskID := ra.generateTaskID(response.RequestData)
	ra.Logger.Info("Beginning signature collection process",
		"taskID", taskID,
		"operatorID", response.OperatorID,
	)

	// Get or create a mutex for this task to protect task initialization
	var taskLock *sync.Mutex
	lockObj, exists := ra.tasksLocks.Load(taskID)
	if !exists {
		// First request for this task - create a new mutex
		taskLock = &sync.Mutex{}
		ra.tasksLocks.Store(taskID, taskLock)
	} else {
		// Use existing mutex for this task
		taskLock = lockObj.(*sync.Mutex)
	}

	// Protect task initialization with the mutex
	taskLock.Lock()
	defer taskLock.Unlock()

	// Try to get existing task
	taskValue, taskExists := ra.tasks.Load(taskID)
	var task *Task

	if !taskExists {
		// Task doesn't exist yet - create a new one
		ra.Logger.Debug("Creating new signature collection task", "taskID", taskID)

		// Extract chain ID and verify configuration exists
		chainID := big.NewInt(response.RequestData.ChainId)
		chainConfig, ok := ra.chainConfigs[chainID.Uint64()]
		if !ok {
			return fmt.Errorf("chain config not found for chain ID: %s", chainID.String())
		}

		// Extract group numbers and thresholds from request
		groupNumbers := types.GroupNumbers{}
		for _, groupNumber := range response.RequestData.GroupNumbers {
			groupNumbers = append(groupNumbers, types.GroupNumber(groupNumber))
		}

		thresholdPercentages := types.GroupThresholdPercentages{}
		for _, thresholdPercentage := range response.RequestData.GroupThresholdPercentages {
			thresholdPercentages = append(thresholdPercentages, types.GroupThresholdPercentage(thresholdPercentage))
		}

		blockNumber := uint32(response.RequestData.Height)

		// Get operator and group state information - required for aggregation
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

		// Create a new task with all required information
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

		// Store the task
		ra.tasks.Store(taskID, task)

		// Set timeout for automatic task finalization
		task.timer = time.AfterFunc(ra.operatorResponseTimeout, func() {
			// Will be executed in a separate goroutine when timeout is reached
			ra.finalizeTask(taskID)
		})

		// Start processing responses in the worker pool
		ra.taskPool.Submit(func() { ra.processResponses(task) })
	} else {
		// Task already exists, use it
		task = taskValue.(*Task)
		ra.Logger.Debug("Using existing signature collection task", "taskID", taskID)
	}

	// Send the response to the task's processing channel
	// This doesn't need to be under the lock because channels are goroutine-safe
	task.responsesChan <- *response

	// Wait for the task to be processed and get the result
	// This blocks until the result is available (either when enough signatures
	// are collected or when timeout occurs)
	validatedResponse := <-task.done

	// Copy the result to the output parameter
	*result = validatedResponse

	ra.Logger.Info("Signature collection process completed",
		"taskID", taskID,
		"operatorID", response.OperatorID,
	)

	return nil
}

// generateTaskID creates a unique identifier for a task based on the request hash
func (ra *RPCServerAggregator) generateTaskID(request avsitypes.DVSRequest) string {
	return hex.EncodeToString(request.Hash())
}

// processResponses continuously processes incoming operator responses for a task
// This runs in a worker from the pool and handles adding responses to the task
func (ra *RPCServerAggregator) processResponses(task *Task) {
	for response := range task.responsesChan {
		// Verify the operator exists in the system
		if _, err := ra.dvsReader.GetOperatorInfoByID(response.OperatorID); err != nil {
			ra.Logger.Error("Failed to get operator info",
				"taskID", task.taskID, "operatorID", response.OperatorID,
				"error", err,
			)
			continue
		}

		// Get the lock for this task
		taskMutex, exists := ra.tasksLocks.Load(task.taskID)
		if !exists {
			return
		}

		// Protect updates to the task with a mutex
		taskMutex.(*sync.Mutex).Lock()

		// Recheck that task still exists after acquiring the lock
		task, exists := ra.tasks.Load(task.taskID)
		if !exists {
			taskMutex.(*sync.Mutex).Unlock()
			return
		}

		// Store the response and update digest mapping
		task.(*Task).operatorResponses[response.OperatorID] = response
		task.(*Task).digestToOperators[response.Digest] = append(task.(*Task).digestToOperators[response.Digest], response.OperatorID)

		taskMutex.(*sync.Mutex).Unlock()
	}
	ra.Logger.Info("Response processing completed", "taskID", task.taskID)
}

// finalizeTask completes a task by aggregating signatures and sending results
// Called either when timeout is reached or when enough signatures are collected
func (ra *RPCServerAggregator) finalizeTask(taskID string) {
	// Get the lock for this task
	taskMutex, exists := ra.tasksLocks.Load(taskID)
	if !exists {
		return
	}

	// Protect the finalization process with a mutex
	taskMutex.(*sync.Mutex).Lock()
	defer taskMutex.(*sync.Mutex).Unlock()

	// Get the task, if it still exists
	task, exists := ra.tasks.Load(taskID)
	if !exists {
		return
	}

	// Stop the timeout timer
	task.(*Task).timer.Stop()

	// Aggregate signatures from all responses
	aggregatedResult, err := ra.aggregateSignatures(task.(*Task))
	if err != nil {
		// Log with more context including the task ID
		ra.Logger.Error("Failed to aggregate signatures",
			"taskID", taskID,
			"error", err)

		// Create a more detailed error response with the original error message
		aggregatedResult = ra.createErrorValidatedResponse(taskID, &rpctypes.RPCError{
			Code:    errcode.AggregationFailed,
			Message: fmt.Sprintf("Failed to aggregate signatures: %v", err),
			Data:    taskID,
		})
	}

	// Send the result to all waiting operators
	for operatorID := range task.(*Task).operatorResponses {
		task.(*Task).done <- *aggregatedResult
		ra.Logger.Info("Sent finalized result to operator", "taskID", taskID, "operatorID", operatorID)
	}

	// Close channels to signal completion
	close(task.(*Task).done)
	close(task.(*Task).responsesChan)

	// Remove the task from the map
	ra.tasks.Delete(taskID)

	ra.Logger.Info("Task finalized and deleted", "taskID", taskID)
}

// createErrorValidatedResponse generates a response object for error cases
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

// aggregateSignatures processes collected signatures from operators and generates the final
// aggregated signature along with necessary validation data for the blockchain.
// The function selects a consensus digest, verifies stake thresholds, and creates
// the final aggregation result.
func (ra *RPCServerAggregator) aggregateSignatures(task *Task) (*aggregator.ValidatedResponse, error) {
	ra.Logger.Info("Starting signature aggregation process", "taskID", task.taskID)

	// Ensure we have signatures to aggregate
	if len(task.operatorResponses) == 0 {
		return nil, errors.New("no signatures to aggregate")
	}

	// STEP 1: Compute total stake per group and prepare threshold calculations
	// Map of total stake for each group
	totalStakePerGroup := make(map[types.GroupNumber]*big.Int)
	for groupNum, groupDvsState := range task.groupOperatorMap {
		totalStakePerGroup[groupNum] = groupDvsState.TotalStake
	}

	// Map of threshold percentages for each group
	thresholdPercentagesMap := make(map[types.GroupNumber]types.GroupThresholdPercentage)
	for groupNum, thresholdPercentage := range task.thresholdPercentages {
		thresholdPercentagesMap[types.GroupNumber(groupNum)] = thresholdPercentage
	}

	// STEP 2: Find a digest that meets the stake threshold requirements
	var selectedDigest ResultDigest
	var selectedData []byte
	var digestSelected bool

	// Iterate through each digest and its signers
	for digest, signerOperatorIDs := range task.digestToOperators {
		// Calculate total stake for these signers across all groups
		signersTotalStakePerGroup := make(map[types.GroupNumber]*big.Int)

		// Process each signer's stake
		for _, operatorID := range signerOperatorIDs {
			// Add this operator's stake to the group totals
			for groupNumber, stake := range task.operatorsDvsStateDict[operatorID].StakePerGroup {
				if _, ok := signersTotalStakePerGroup[groupNumber]; !ok {
					signersTotalStakePerGroup[groupNumber] = big.NewInt(0)
				}
				signersTotalStakePerGroup[groupNumber].Add(signersTotalStakePerGroup[groupNumber], stake)
			}
		}

		// Check if this digest has enough stake to meet thresholds
		if ra.checkIfStakeThresholdsMet(signersTotalStakePerGroup, totalStakePerGroup, thresholdPercentagesMap) {
			selectedDigest = digest
			selectedData = task.operatorResponses[signerOperatorIDs[0]].Data
			digestSelected = true
			ra.Logger.Debug("Selected digest meets stake thresholds", "digest", selectedDigest)
			break
		}
	}

	// If no digest met the threshold requirements, return an error
	if !digestSelected {
		return nil, fmt.Errorf("no digest meets stake thresholds")
	}

	// STEP 3: Aggregate signatures from operators that signed the selected digest

	// Initialize aggregated signature and signers' aggregate public key
	aggregatedSignature := bls.NewZeroSignature()
	signersApkG2 := bls.NewZeroG2Point()

	// Collect all group aggregate public keys (G1 points)
	groupApksG1 := make([]*bls.G1Point, 0, len(task.groupNumbers))
	for _, groupNumber := range task.groupNumbers {
		groupApksG1 = append(groupApksG1, task.groupOperatorMap[groupNumber].AggPubkeyG1)
	}

	// Add signatures and public keys for operators that signed the selected digest
	for operatorID, response := range task.operatorResponses {
		if response.Digest == selectedDigest {
			operator, _ := ra.dvsReader.GetOperatorInfoByID(operatorID)
			aggregatedSignature.Add(response.Signature)
			signersApkG2.Add(operator.Pubkeys.G2Pubkey)
		}
	}

	// STEP 4: Identify non-signers for the selected digest
	// Identify active groups
	activeGroups := make(map[types.GroupNumber]bool)

	// Build maps of registered operators
	registeredOperators := make(map[types.OperatorID]types.OperatorInfo)
	operatorMapByBLSOperatorID := make(map[types.OperatorID]types.OperatorInfo)

	// Gather operator information for each group
	for groupNum, operatorInfos := range task.operatorStateInfo.GroupOperatorMap {
		if len(operatorInfos) > 0 {
			activeGroups[groupNum] = true
		}

		for _, operator := range operatorInfos {
			addrOperatorID := operator.OperatorID
			operatorInfo, err := ra.dvsReader.GetOperatorInfoByID(addrOperatorID)
			if err != nil {
				return nil, fmt.Errorf("failed to get operator info by ID: %w", err)
			}

			registeredOperators[addrOperatorID] = operatorInfo
			blsOperatorID := operatorInfo.Pubkeys.GetOperatorID()
			operatorMapByBLSOperatorID[blsOperatorID] = operatorInfo
		}
	}

	// Identify operators that did not sign or signed a different digest
	nonSignersOperatorIds := make([]types.OperatorID, 0)
	for addrOperatorID, operatorInfo := range registeredOperators {
		// Check if operator didn't sign or signed a different digest
		response, signed := task.operatorResponses[addrOperatorID]
		if !signed || response.Digest != selectedDigest {
			// Check if they're in an active group
			isInActiveGroup := false
			for _, groupNum := range task.groupNumbers {
				if activeGroups[types.GroupNumber(groupNum)] {
					isInActiveGroup = true
					break
				}
			}

			if isInActiveGroup {
				blsOperatorID := operatorInfo.Pubkeys.GetOperatorID()
				nonSignersOperatorIds = append(nonSignersOperatorIds, blsOperatorID)
			}
		}
	}

	// Sort non-signers by operator ID for deterministic output
	sort.SliceStable(nonSignersOperatorIds, func(i, j int) bool {
		iOprInt := new(big.Int).SetBytes(nonSignersOperatorIds[i][:])
		jOprInt := new(big.Int).SetBytes(nonSignersOperatorIds[j][:])
		return iOprInt.Cmp(jOprInt) == -1
	})

	// Build list of non-signer public keys
	nonSignersPubkeysG1 := make([]*bls.G1Point, 0, len(nonSignersOperatorIds))
	for _, blsOperatorID := range nonSignersOperatorIds {
		operator, ok := operatorMapByBLSOperatorID[blsOperatorID]
		if !ok {
			ra.Logger.Error("Operator not found by blsOperatorID", "blsOperatorID", blsOperatorID)
			continue
		}
		nonSignersPubkeysG1 = append(nonSignersPubkeysG1, operator.Pubkeys.G1Pubkey)
	}

	// STEP 5: Get indices for on-chain verification
	indices, err := ra.dvsReader.GetCheckSignaturesIndices(
		task.chainConfig.ChainID,
		task.blockNumber,
		task.groupNumbers,
		nonSignersOperatorIds,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get check signatures indices: %w", err)
	}

	// STEP 6: Create and return the final validated response
	ra.Logger.Info("Signature aggregation completed successfully",
		"taskID", task.taskID,
		"signerCount", len(task.operatorResponses)-len(nonSignersOperatorIds),
		"nonSignerCount", len(nonSignersOperatorIds),
	)

	// Construct the final validated response with all necessary data
	result := &aggregator.ValidatedResponse{
		Err:                         nil,
		Hash:                        []byte(task.taskID),
		Data:                        selectedData,
		NonSignersPubkeysG1:         nonSignersPubkeysG1,
		GroupApksG1:                 groupApksG1,
		SignersApkG2:                signersApkG2,
		SignersAggSigG1:             aggregatedSignature,
		NonSignerGroupBitmapIndices: indices.NonSignerGroupBitmapIndices,
		GroupApkIndices:             indices.GroupApkIndices,
		TotalStakeIndices:           indices.TotalStakeIndices,
		NonSignerStakeIndices:       indices.NonSignerStakeIndices,
	}

	return result, nil
}

// checkIfStakeThresholdsMet determines if the signatures collected have enough
// stake to meet the required thresholds for each group
func (ra *RPCServerAggregator) checkIfStakeThresholdsMet(
	signedStakePerGroup map[types.GroupNumber]*big.Int,
	totalStakePerGroup map[types.GroupNumber]*big.Int,
	groupThresholdPercentagesMap map[types.GroupNumber]types.GroupThresholdPercentage,
) bool {
	for groupNum, groupThresholdPercentage := range groupThresholdPercentagesMap {
		// Check if we have stake information for this group
		signedStakeByGroup, ok := signedStakePerGroup[groupNum]
		if !ok {
			return false
		}

		totalStakeByGroup, ok := totalStakePerGroup[groupNum]
		if !ok {
			return false
		}

		// Calculate if we meet the threshold percentage
		// signedStake * 100 >= totalStake * thresholdPercentage
		signedStake := new(big.Int).Mul(signedStakeByGroup, big.NewInt(100))
		thresholdStake := new(big.Int).Mul(totalStakeByGroup, big.NewInt(int64(groupThresholdPercentage)))

		if signedStake.Cmp(thresholdStake) < 0 {
			return false
		}
	}
	return true
}

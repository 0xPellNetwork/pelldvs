package security

import (
	"context"
	"fmt"
	"math/big"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cosmos/gogoproto/proto"

	"github.com/0xPellNetwork/pelldvs-interactor/interactor/reader"
	evmtypes "github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	aggtypes "github.com/0xPellNetwork/pelldvs/aggregator"
	avsitypes "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/proxy"
	"github.com/0xPellNetwork/pelldvs/state/requestindex"
	"github.com/0xPellNetwork/pelldvs/types"
)

const (
	responseDigestLenLimit = 32
)

type DVSReactor struct {
	config            config.PellConfig
	ProxyApp          proxy.AppConns
	dvsState          *DVSState
	logger            log.Logger
	dvsRequestIndexer requestindex.DvsRequestIndexer
	dvsReader         reader.DVSReader
	eventManager      *EventManager
}

// CreateDVSReactor creates a new DVSReactor instance
func CreateDVSReactor(
	config config.PellConfig,
	proxyApp proxy.AppConns,
	dvsRequestIndexer requestindex.DvsRequestIndexer,
	db dbm.DB,
	dvsState *DVSState,
	logger log.Logger,
	eventManager *EventManager,
) (DVSReactor, error) {

	dvsReader, err := reader.NewDVSReader(config.InteractorConfigPath, db, logger)
	if err != nil {
		return DVSReactor{}, fmt.Errorf("failed to create dvsReader: %v", err)
	}

	dvs := DVSReactor{
		config:            config,
		ProxyApp:          proxyApp,
		dvsState:          dvsState,
		logger:            logger,
		dvsRequestIndexer: dvsRequestIndexer,
		dvsReader:         dvsReader,
		eventManager:      eventManager,
	}
	return dvs, nil
}

// SaveDVSRequestResult saves the DVS request result
func (dvs *DVSReactor) SaveDVSRequestResult(res *avsitypes.DVSRequestResult, first bool) error {
	if first {
		rawBytesDvsRequest, err := proto.Marshal(res.DvsRequest)
		if err != nil {
			return err
		}
		hash := types.DvsRequest(rawBytesDvsRequest).Hash()
		old, err := dvs.dvsRequestIndexer.Get(hash)
		if err != nil {
			return err
		}
		if old != nil {
			return fmt.Errorf("DVS request hash %X already exist", hash)
		}
	}
	return dvs.dvsRequestIndexer.Index(res)
}

// HandleDVSRequest handles the DVS request
func (dvs *DVSReactor) HandleDVSRequest(request avsitypes.DVSRequest) error {
	dvs.logger.Info("dvsReactor.HandleDVSRequest", "request", request)

	// First save the request
	result := avsitypes.DVSRequestResult{
		DvsRequest: &request,
	}
	if err := dvs.SaveDVSRequestResult(&result, true); err != nil {
		dvs.logger.Error("dvsReactor dvsindex.Index", "err", err.Error())
		return err
	}

	groupNumbers := make(evmtypes.GroupNumbers, len(request.GroupNumbers))
	for i, v := range request.GroupNumbers {
		groupNumbers[i] = evmtypes.GroupNumber(v)
	}
	operatorsDvsState, err := dvs.dvsReader.GetOperatorsDVSStateAtBlock(uint64(request.ChainId), groupNumbers, uint32(request.Height))
	if err != nil {
		dvs.logger.Error("dvsInteractor dvsReader.GetOperatorsDVSStateAtBlock", "err", err.Error())
		return err
	}

	operators := make([]*avsitypes.Operator, 0)
	for _, operatorState := range operatorsDvsState {
		stake := big.NewInt(0)
		for _, stakeAmount := range operatorState.StakePerGroup {
			stake = stake.Add(stake, stakeAmount)
		}
		pubkeys := &avsitypes.OperatorPubkeys{
			G1Pubkey: operatorState.OperatorInfo.Pubkeys.G1Pubkey.Serialize(),
			G2Pubkey: operatorState.OperatorInfo.Pubkeys.G2Pubkey.Serialize(),
		}
		operators = append(operators, &avsitypes.Operator{
			Id:      operatorState.OperatorID[:],
			Address: operatorState.OperatorAddress[:],
			MetaUri: operatorState.OperatorInfo.MetaURI.String(),
			Socket:  operatorState.OperatorInfo.Socket.String(),
			Stake:   stake.Int64(),
			Pubkeys: pubkeys,
		})
	}

	response, err := dvs.ProxyApp.Dvs().ProcessDVSRequest(context.Background(), &avsitypes.RequestProcessDVSRequest{
		Request:  &request,
		Operator: operators,
	})
	if err != nil {
		dvs.logger.Error("dvsReactor pellProxyApp.ProcessDVSRequest", "err", err.Error())
		return err
	}

	// Check if responseDigest length is equal to 32
	if len(response.ResponseDigest) != responseDigestLenLimit {
		dvs.logger.Error("responseDigest length is not equal to 32", "responseDigest", response.ResponseDigest)
		return fmt.Errorf("responseDigest length %d is not equal to %d", response.ResponseDigest, responseDigestLenLimit)
	}

	// Second save the request
	result.ResponseProcessDvsRequest = response
	if err := dvs.SaveDVSRequestResult(&result, false); err != nil {
		dvs.logger.Error("dvsReactor dvsindex.Index", "err", err.Error())
		return err
	}

	dvs.eventManager.eventBus.Publish(types.CollectResponseSignatureRequest, request.Hash())
	return nil
}

// OnRequestAfterAggregated is called after the request is aggregated
func (dvs *DVSReactor) OnRequestAfterAggregated(requestHash avsitypes.DVSRequestHash, validatedResponse aggtypes.ValidatedResponse) error {
	dvs.logger.Info("dvsReactor.OnRequestAfterAggregated", "requestHash", requestHash)

	if validatedResponse.Err != nil {
		dvs.logger.Error("Aggregator returned an error", "error", validatedResponse.Err)
		return validatedResponse.Err
	}

	// Build dvs response
	publicG1 := make([][]byte, 0, len(validatedResponse.NonSignersPubkeysG1))
	for _, v := range validatedResponse.NonSignersPubkeysG1 {
		publicG1 = append(publicG1, v.Serialize())
	}

	apksG1 := make([][]byte, 0, len(validatedResponse.GroupApksG1))
	for _, v := range validatedResponse.GroupApksG1 {
		apksG1 = append(apksG1, v.Serialize())
	}

	indices := make([]*avsitypes.NonSignerStakeIndice, 0, len(validatedResponse.NonSignerStakeIndices))
	for _, v := range validatedResponse.NonSignerStakeIndices {
		indices = append(indices, &avsitypes.NonSignerStakeIndice{
			NonSignerStakeIndice: v,
		})
	}

	dvsResponse := avsitypes.DVSResponse{
		Data:                        validatedResponse.Data,
		Hash:                        validatedResponse.Hash,
		NonSignersPubkeysG1:         publicG1,
		GroupApksG1:                 apksG1,
		SignersApkG2:                validatedResponse.SignersApkG2.Serialize(),
		SignersAggSigG1:             validatedResponse.SignersAggSigG1.Serialize(),
		NonSignerGroupBitmapIndices: validatedResponse.NonSignerGroupBitmapIndices,
		GroupApkIndices:             validatedResponse.GroupApkIndices,
		TotalStakeIndices:           validatedResponse.TotalStakeIndices,
		NonSignerStakeIndices:       indices,
	}

	// Query request result
	result, err := dvs.dvsRequestIndexer.Get(requestHash)
	if err != nil {
		dvs.logger.Error("dvsReactor.dvsindex.Get", "err", err.Error())
		return err
	}

	// Third save the request
	result.DvsResponse = &dvsResponse
	if err := dvs.SaveDVSRequestResult(result, false); err != nil {
		dvs.logger.Error("dvsReactor.dvsindex.Index dvsResponseIdx", "err", err.Error())
		return err
	}

	// If no error, send validated response to proxy application
	postResponse := &avsitypes.RequestProcessDVSResponse{
		DvsResponse: &dvsResponse,
		DvsRequest:  result.DvsRequest,
	}
	responseProcessDVSResponse, err := dvs.ProxyApp.Dvs().ProcessDVSResponse(context.Background(), postResponse)
	if err != nil {
		return err
	}

	// Fourth save the request
	result.ResponseProcessDvsResponse = responseProcessDVSResponse
	if err := dvs.SaveDVSRequestResult(result, false); err != nil {
		dvs.logger.Error("dvsReactor.dvsindex.Index dvsResponseIdx", "err", err.Error())
		return err
	}

	// Log validated response details
	dvs.logger.Info("Validated Response Details",
		"Hash", validatedResponse.Hash,
		"NonSignerGroupBitmapIndices", validatedResponse.NonSignerGroupBitmapIndices,
		"NonSignersPubkeysG1", validatedResponse.NonSignersPubkeysG1,
		"GroupApksG1", validatedResponse.GroupApksG1,
		"SignersApkG2", validatedResponse.SignersApkG2,
		"SignersAggSigG1", validatedResponse.SignersAggSigG1,
		"GroupApkIndices", validatedResponse.GroupApkIndices,
		"TotalStakeIndices", validatedResponse.TotalStakeIndices,
		"NonSignerStakeIndices", validatedResponse.NonSignerStakeIndices,
	)
	return nil
}

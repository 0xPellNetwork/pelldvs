package security

import (
	"context"
	"fmt"
	"math/big"
	"time"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cosmos/gogoproto/proto"

	"github.com/0xPellNetwork/pelldvs-interactor/interactor/reader"
	evmtypes "github.com/0xPellNetwork/pelldvs-interactor/types"
	"github.com/0xPellNetwork/pelldvs-libs/crypto/bls"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	aggtypes "github.com/0xPellNetwork/pelldvs/aggregator"
	avsitypes "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/proxy"
	"github.com/0xPellNetwork/pelldvs/state/requestindex"
	"github.com/0xPellNetwork/pelldvs/types"
)

type DVSReactor struct {
	config            config.PellConfig
	ProxyApp          proxy.AppConns
	dvsState          *DVSState
	logger            log.Logger
	aggregator        aggtypes.Aggregator
	dvsRequestIndexer requestindex.DvsRequestIndexer
	dvsReader         reader.DVSReader
}

func CreateDVSReactor(
	config config.PellConfig,
	proxyApp proxy.AppConns,
	aggregator aggtypes.Aggregator,
	storeDir string,
	dvsRequestIndexer requestindex.DvsRequestIndexer,
	db dbm.DB,
	logger log.Logger,
) (DVSReactor, error) {
	dvsReqStore, err := NewStore(storeDir)
	if err != nil {
		return DVSReactor{}, fmt.Errorf("failed to create DVSReqStore: %v", err)
	}

	dvsState, err := NewDVSState(&config, dvsReqStore, storeDir)
	if err != nil {
		return DVSReactor{}, fmt.Errorf("failed to create DVSState: %v", err)
	}

	dvsReader, err := reader.NewDVSReader(config.InteractorConfigPath, db, logger)
	if err != nil {
		return DVSReactor{}, fmt.Errorf("failed to create dvsReader: %v", err)
	}

	dvs := DVSReactor{
		config:            config,
		ProxyApp:          proxyApp,
		dvsState:          dvsState,
		logger:            logger,
		aggregator:        aggregator,
		dvsRequestIndexer: dvsRequestIndexer,
		dvsReader:         dvsReader,
	}

	return dvs, nil
}

func (dvs *DVSReactor) SignMessage(message []byte) (*bls.Signature, error) {
	return dvs.dvsState.privValidator.SignMessage(message)
}

func (dvs *DVSReactor) OnQuery(key []byte) ([]byte, []byte, error) {
	res, err := dvs.ProxyApp.Query().Query(context.Background(), &avsitypes.RequestQuery{
		Data: key,
	})
	if err != nil {
		return nil, nil, err
	}
	return res.Key, res.Value, nil
}

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

func (dvs *DVSReactor) OnRequest(request avsitypes.DVSRequest) (*avsitypes.DVSRequestResult, error) {

	reqIdx := avsitypes.DVSRequestResult{
		DvsRequest: &request,
	}

	if err := dvs.SaveDVSRequestResult(&reqIdx, true); err != nil {
		dvs.logger.Error("dvs.SaveDVSRequest save req", "err", err.Error())
		return nil, err
	}

	groupNumbers := make(evmtypes.GroupNumbers, len(request.GroupNumbers))
	for i, v := range request.GroupNumbers {
		groupNumbers[i] = evmtypes.GroupNumber(v)
	}
	operatorsDvsState, err := dvs.dvsReader.GetOperatorsDVSStateAtBlock(uint64(request.ChainId), groupNumbers, uint32(request.Height))
	if err != nil {
		dvs.logger.Error("dvsInteractor.GetOperatorsDVSStateAtBlock", "err", err.Error())
		return nil, err
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

	responseProcessDVSRequest, err := dvs.ProxyApp.Dvs().ProcessDVSRequest(context.Background(), &avsitypes.RequestProcessDVSRequest{
		Request:  &request,
		Operator: operators,
	})

	if err != nil {
		return nil, err
	}

	reqResIdx := avsitypes.DVSRequestResult{
		DvsRequest:                &request,
		ResponseProcessDvsRequest: responseProcessDVSRequest,
	}

	if err := dvs.SaveDVSRequestResult(&reqResIdx, false); err != nil {
		dvs.logger.Error("dvs.dvsindex.Index", "err", err.Error())
		return nil, err
	}

	signature, err := dvs.SignMessage(responseProcessDVSRequest.ResponseDigest)
	if err != nil {
		dvs.logger.Error("SignMessage failed", "error", err)
		return nil, err
	}
	dvs.logger.Debug("responseWithSignature", "signature", signature)

	var digestArr [32]byte
	copy(digestArr[:], responseProcessDVSRequest.ResponseDigest)
	g1p := bls.G1Point{
		G1Affine: signature.G1Affine,
	}

	sig := bls.Signature{G1Point: &g1p}

	responseWithSingature := aggtypes.ResponseWithSignature{
		Data:        responseProcessDVSRequest.Response,
		Signature:   &sig,
		OperatorID:  dvs.dvsState.operatorID,
		RequestData: request,
		Digest:      digestArr,
	}

	// Create a channel to receive validated response
	validatedResponseCh := make(chan aggtypes.ValidatedResponse, 1)

	// Send response signature to aggregator and wait for result
	err = dvs.aggregator.CollectResponseSignature(&responseWithSingature, validatedResponseCh)
	if err != nil {
		dvs.logger.Error("Failed to send response signature to aggregator", "error", err)
		return nil, fmt.Errorf("failed to send response signature to aggregator: %v", err)
	}

	var responseProcessDVSResponse *avsitypes.ResponseProcessDVSResponse
	// Wait for validated response
	select {
	case validatedResponse := <-validatedResponseCh:

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

		if validatedResponse.Err == nil {
			dvsRes := avsitypes.DVSResponse{
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

			dvsResponseIdx := avsitypes.DVSRequestResult{
				DvsRequest:                &request,
				ResponseProcessDvsRequest: responseProcessDVSRequest,
				DvsResponse:               &dvsRes,
			}

			if err := dvs.SaveDVSRequestResult(&dvsResponseIdx, false); err != nil {
				dvs.logger.Error("dvs.dvsindex.Index dvsResponseIdx", "err", err.Error())
				return nil, err
			}

			// If no error, send validated response to proxy application
			postReq := &avsitypes.RequestProcessDVSResponse{
				DvsResponse: &dvsRes,
				DvsRequest:  &request,
			}

			responseProcessDVSResponse, err = dvs.ProxyApp.Dvs().ProcessDVSResponse(context.Background(), postReq)
			if err != nil {
				return nil, err
			}

			resultIdx := avsitypes.DVSRequestResult{
				DvsRequest:                 &request,
				ResponseProcessDvsRequest:  responseProcessDVSRequest,
				DvsResponse:                &dvsRes,
				ResponseProcessDvsResponse: responseProcessDVSResponse,
			}

			if err := dvs.SaveDVSRequestResult(&resultIdx, false); err != nil {
				dvs.logger.Error("dvs.dvsindex.Index dvsResponseIdx", "err", err.Error())
				return nil, err
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

		} else {
			dvs.logger.Error("Aggregator returned an error", "error", validatedResponse.Err)
		}
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for aggregator response")
	}

	dvs.logger.Debug("responseWithSingature", "responseWithSingature", responseWithSingature)
	dvs.logger.Debug("responseProcessDVSRequest", "responseProcessDVSRequest", responseProcessDVSRequest)
	dvs.logger.Debug("pellProxyApp", "pellProxyApp", dvs.ProxyApp)
	dvs.logger.Debug("OnRequest", "request", request)

	return &avsitypes.DVSRequestResult{
		DvsRequest:                 &request,
		ResponseProcessDvsRequest:  responseProcessDVSRequest,
		ResponseProcessDvsResponse: responseProcessDVSResponse,
	}, nil
}

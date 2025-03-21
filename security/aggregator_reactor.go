package security

import (
	"fmt"
	"github.com/0xPellNetwork/pelldvs-libs/crypto/bls"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	aggtypes "github.com/0xPellNetwork/pelldvs/aggregator"
	avsitypes "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/state/requestindex"
	"github.com/0xPellNetwork/pelldvs/types"
)

type AggregatorReactor struct {
	aggregator        aggtypes.Aggregator
	dvsRequestIndexer requestindex.DvsRequestIndexer
	privValidator     types.PrivValidator
	dvsState          *DVSState
	logger            log.Logger
	eventManager      *EventManager
}

func CreateAggregatorReactor(
	aggregator aggtypes.Aggregator,
	dvsRequestIndexer requestindex.DvsRequestIndexer,
	privValidator types.PrivValidator,
	dvsState *DVSState,
	logger log.Logger,
	eventManager *EventManager,
) *AggregatorReactor {

	return &AggregatorReactor{
		aggregator:        aggregator,
		dvsRequestIndexer: dvsRequestIndexer,
		privValidator:     privValidator,
		dvsState:          dvsState,
		logger:            logger,
		eventManager:      eventManager,
	}
}

type AggregatorResponse struct {
	requestHash      avsitypes.DVSRequestHash
	validateResponse aggtypes.ValidatedResponse
}

func (ar *AggregatorReactor) HandleSignatureCollectionRequest(requestHash avsitypes.DVSRequestHash) error {
	ar.logger.Info("HandleSignatureCollectionRequest", "requestHash", requestHash)

	result, err := ar.dvsRequestIndexer.Get(requestHash)
	if err != nil {
		ar.logger.Error("AggregatorReactor: Get request from indexer failed", "error", err)
		return err
	}

	response := result.ResponseProcessDvsRequest
	signature, err := ar.privValidator.SignBytes(response.ResponseDigest)
	if err != nil {
		ar.logger.Error("SignMessage failed", "error", err)
		return err
	}
	ar.logger.Debug("responseWithSignature", "signature", signature)

	g1p := bls.G1Point{
		G1Affine: signature.G1Affine,
	}
	sig := bls.Signature{G1Point: &g1p}

	responseWithSignature := aggtypes.ResponseWithSignature{
		Data:        response.Response,
		Signature:   &sig,
		OperatorID:  ar.dvsState.operatorID,
		RequestData: *result.DvsRequest,
		Digest:      [32]byte(response.ResponseDigest),
	}

	// Create a channel to receive validated response
	validatedResponseCh := make(chan aggtypes.ValidatedResponse, 1)

	// Send response signature to aggregatorReactor and wait for result
	err = ar.aggregator.CollectResponseSignature(&responseWithSignature, validatedResponseCh)
	if err != nil {
		ar.logger.Error("Failed to send response signature to aggregatorReactor", "error", err)
		return fmt.Errorf("failed to send response signature to aggregatorReactor: %v", err)
	}

	aggregatedResponse := AggregatorResponse{
		requestHash:      requestHash,
		validateResponse: <-validatedResponseCh,
	}

	ar.eventManager.eventBus.Publish(types.CollectResponseSignatureDone, aggregatedResponse)
	return nil
}

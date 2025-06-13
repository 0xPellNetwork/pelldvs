package security

import (
	"fmt"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/bls"
	"github.com/0xPellNetwork/pelldvs-libs/log"
	aggtypes "github.com/0xPellNetwork/pelldvs/aggregator/types"
	avsitypes "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/state/requestindex"
	"github.com/0xPellNetwork/pelldvs/types"
)

// AggregatorReactor handles the collection and aggregation of response signatures
// from operators, interfacing between the DVS system and the aggregator service
type AggregatorReactor struct {
	aggClient         aggtypes.Aggregator
	dvsRequestIndexer requestindex.DvsRequestIndexer
	privValidator     types.PrivValidator
	dvsState          *DVSState
	logger            log.Logger
	eventManager      *EventManager
}

// CreateAggregatorReactor initializes a new AggregatorReactor with all required dependencies
// to handle signature collection and aggregation operations
func CreateAggregatorReactor(
	aggClient aggtypes.Aggregator,
	dvsRequestIndexer requestindex.DvsRequestIndexer,
	privValidator types.PrivValidator,
	dvsState *DVSState,
	logger log.Logger,
	eventManager *EventManager,
) *AggregatorReactor {
	return &AggregatorReactor{
		aggClient:         aggClient,
		dvsRequestIndexer: dvsRequestIndexer,
		privValidator:     privValidator,
		dvsState:          dvsState,
		logger:            logger,
		eventManager:      eventManager,
	}
}

// AggregatorResponse encapsulates the result of an aggregation operation,
// containing both the original request hash and the validated response
type AggregatorResponse struct {
	requestHash      avsitypes.DVSRequestHash
	validateResponse aggtypes.ValidatedResponse
}

// HandleSignatureCollectionRequest processes a signature collection request
// by retrieving the original request, signing the response, and submitting it
// to the aggregator for collection and validation
func (ar *AggregatorReactor) HandleSignatureCollectionRequest(requestHash avsitypes.DVSRequestHash) error {
	ar.logger.Info("HandleSignatureCollectionRequest", "requestHash", requestHash)

	// recover from panic
	defer func() {
		if r := recover(); r != nil {
			ar.logger.Error("HandleSignatureCollectionRequest panic",
				"requestHash", requestHash,
				"error", fmt.Sprintf("%v", r),
			)
		}
	}()

	// Get request from indexer
	result, err := ar.dvsRequestIndexer.Get(requestHash)
	if err != nil {
		ar.logger.Error("AggregatorReactor: Get request from indexer failed", "error", err)
		return err
	}

	// Extract the response and sign its digest
	response := result.ResponseProcessDvsRequest
	signature, err := ar.privValidator.SignBytes(response.ResponseDigest)
	if err != nil {
		ar.logger.Error("SignMessage failed", "error", err)
		return err
	}
	ar.logger.Debug("responseWithSignature", "signature", signature)

	// Convert the signature to the required BLS format
	sig := bls.Signature{G1Point: &bls.G1Point{
		G1Affine: signature.G1Affine,
	}}

	// Create a response with signature object for aggregation
	responseWithSignature := aggtypes.ResponseWithSignature{
		Data:        response.Response,
		Signature:   &sig,
		OperatorID:  ar.dvsState.operatorID,
		RequestData: *result.DvsRequest,
		Digest:      [32]byte(response.ResponseDigest),
	}

	// Create a channel to receive validated response
	validatedResponseCh := make(chan aggtypes.ValidatedResponse, 1)

	ar.logger.Info("HandleSignatureCollectionRequest, call CollectResponseSignature",
		"responseWithSignature", responseWithSignature,
	)
	// Send response signature to aggregator and wait for result
	if err = ar.aggClient.CollectResponseSignature(&responseWithSignature, validatedResponseCh); err != nil {
		ar.logger.Error("Failed to send response signature to aggregator", "error", err)
		return fmt.Errorf("failed to send response signature to aggregator: %v", err)
	}
	ar.logger.Info("HandleSignatureCollectionRequest, CollectResponseSignature done")

	// Create an aggregator response with the validated result
	aggregatedResponse := AggregatorResponse{
		requestHash:      requestHash,
		validateResponse: <-validatedResponseCh,
	}

	ar.logger.Info("HandleSignatureCollectionRequest",
		"aggregatedResponse.requestHash", aggregatedResponse.requestHash,
		"aggregatedResponse.validateResponse", aggregatedResponse.validateResponse,
	)

	// Publish the completion event with the aggregated response
	ar.eventManager.eventBus.Pub(types.CollectResponseSignatureDone, aggregatedResponse)

	ar.logger.Info("HandleSignatureCollectionRequest done, event sent")
	return nil
}

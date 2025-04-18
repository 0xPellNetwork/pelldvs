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

// AggregatorReactor handles the process of collecting and aggregating signatures
// for distributed validation system (DVS) requests
type AggregatorReactor struct {
	aggregator        aggtypes.Aggregator            // Performs signature aggregation
	dvsRequestIndexer requestindex.DvsRequestIndexer // Indexes and retrieves DVS requests
	privValidator     types.PrivValidator            // Signs messages with node's private key
	dvsState          *DVSState                      // Maintains DVS operational state
	logger            log.Logger                     // Handles logging
	eventManager      *EventManager                  // Manages event publishing
}

// NewAggregatorReactor initializes and returns a new AggregatorReactor instance
// with all necessary dependencies for signature collection and aggregation
func NewAggregatorReactor(
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

// AggregatorResponse encapsulates the result of a signature aggregation process,
// containing both the original request hash and the validated response
type AggregatorResponse struct {
	requestHash      avsitypes.DVSRequestHash   // Hash of the original DVS request
	validateResponse aggtypes.ValidatedResponse // Aggregated and validated response
}

// HandleSignatureCollectionRequest processes a signature collection request by retrieving
// the original request, signing it, and submitting it to the aggregator
func (ar *AggregatorReactor) HandleSignatureCollectionRequest(requestHash avsitypes.DVSRequestHash) error {
	ar.logger.Info("Processing signature collection request", "requestHash", requestHash)

	// Retrieve the original request from the indexer
	result, err := ar.dvsRequestIndexer.Get(requestHash)
	if err != nil {
		ar.logger.Error("Failed to retrieve request from indexer", "error", err)
		return fmt.Errorf("request retrieval failed: %w", err)
	}

	// Extract response data and generate signature
	response := result.ResponseProcessDvsRequest
	signature, err := ar.privValidator.SignBytes(response.ResponseDigest)
	if err != nil {
		ar.logger.Error("Failed to sign response digest", "error", err)
		return fmt.Errorf("signature generation failed: %w", err)
	}

	// Create BLS signature from raw signature data
	blsSig := &bls.Signature{
		G1Point: &bls.G1Point{
			G1Affine: signature.G1Affine,
		},
	}

	// Package response with signature and metadata
	responseWithSignature := aggtypes.ResponseWithSignature{
		Data:        response.Response,
		Signature:   blsSig,
		OperatorID:  ar.dvsState.operatorID,
		RequestData: *result.DvsRequest,
		Digest:      [32]byte(response.ResponseDigest),
	}

	// Set up async communication channel for aggregator result
	validatedResponseCh := make(chan aggtypes.ValidatedResponse, 1)

	// Submit signature to aggregator for collection and processing
	if err := ar.aggregator.CollectResponseSignature(&responseWithSignature, validatedResponseCh); err != nil {
		ar.logger.Error("Failed to submit signature to aggregator", "error", err)
		return fmt.Errorf("signature submission failed: %w", err)
	}

	// Receive validation result and publish completion event
	aggregatedResponse := AggregatorResponse{
		requestHash:      requestHash,
		validateResponse: <-validatedResponseCh,
	}
	ar.eventManager.eventBus.Publish(types.CollectResponseSignatureDone, aggregatedResponse)

	ar.logger.Debug("Signature collection request completed successfully", "requestHash", requestHash)
	return nil
}

// Package security provides types and functionality for handling security-related
// operations in the distributed validation system
package security

import (
	"github.com/0xPellNetwork/pelldvs/aggregator/types"
	avsitypes "github.com/0xPellNetwork/pelldvs/avsi/types"
)

// RequestProcessRequest encapsulates a DVS request that needs to be processed
// by the validation system
type RequestProcessRequest struct {
	Request avsitypes.DVSRequest
}

// ResponseProcessRequest contains the processed response data and its digest
// following validation of a DVS request
type ResponseProcessRequest struct {
	Response       []byte
	ResponseDigest []byte
}

// ResponsePostRequest represents a confirmation request after a response
// has been successfully posted to the network
type ResponsePostRequest struct{}

// RequestPostRequest contains the validated aggregated response
// to be posted back to the requestor
type RequestPostRequest struct {
	Response types.ValidatedResponse
}

// DVSReqResponse represents the complete lifecycle of a DVS request,
// including the original request, its validated response, and confirmation receipt
type DVSReqResponse struct {
	Request  avsitypes.DVSRequest
	Response *types.ValidatedResponse
	Receipt  *ResponsePostRequest
}

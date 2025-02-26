package security

import (
	"github.com/0xPellNetwork/pelldvs/aggregator"
	avsitypes "github.com/0xPellNetwork/pelldvs/avsi/types"
)

type RequestProcessRequest struct {
	Request avsitypes.DVSRequest
}

type ResponseProcessRequest struct {
	Response       []byte
	ResponseDigest []byte
}

type ResponsePostRequest struct {
}

type RequestPostRequest struct {
	Response aggregator.ValidatedResponse
}

type DVSReqResponse struct {
	Request  avsitypes.DVSRequest
	Response *aggregator.ValidatedResponse
	Receipt  *ResponsePostRequest
}

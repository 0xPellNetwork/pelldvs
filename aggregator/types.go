package aggregator

import (
	"github.com/0xPellNetwork/pelldvs-libs/crypto/bls"
	avsitypes "github.com/0xPellNetwork/pelldvs/avsi/types"
	rpctypes "github.com/0xPellNetwork/pelldvs/rpc/jsonrpc/types"
)

// Aggregator defines the interface for a component that can collect and aggregate
// signatures from distributed validator operators
type Aggregator interface {
	// CollectResponseSignature collects a signature from an operator and returns the validation result
	CollectResponseSignature(response *ResponseWithSignature, result chan<- ValidatedResponse) error
}

// ResponseWithSignature encapsulates an operator's response to a DVS request,
// including the signature and necessary metadata for verification
type ResponseWithSignature struct {
	Data        []byte
	Digest      [32]byte
	Signature   *bls.Signature
	OperatorID  [32]byte
	RequestData avsitypes.DVSRequest
}

// ValidatedResponse represents the result of the signature aggregation process,
// containing either the aggregated signature data or an error if validation failed
type ValidatedResponse struct {
	Data                        []byte
	Err                         *rpctypes.RPCError
	Hash                        []byte
	NonSignersPubkeysG1         []*bls.G1Point
	GroupApksG1                 []*bls.G1Point
	SignersApkG2                *bls.G2Point
	SignersAggSigG1             *bls.Signature
	NonSignerGroupBitmapIndices []uint32
	GroupApkIndices             []uint32
	TotalStakeIndices           []uint32
	NonSignerStakeIndices       [][]uint32
}

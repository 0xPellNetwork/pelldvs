package types

import (
	"github.com/0xPellNetwork/pelldvs-libs/crypto/bls"
	avsitypes "github.com/0xPellNetwork/pelldvs/avsi/types"
	rpctypes "github.com/0xPellNetwork/pelldvs/rpc/jsonrpc/types"
)

// Aggregator defines the interface for signature collection and aggregation
// services in the distributed validation system. It handles the process of
// collecting signatures from operators and aggregating them into a single
// validated response.
type Aggregator interface {
	// CollectResponseSignature collects the response signature from the operator
	CollectResponseSignature(response *ResponseWithSignature, result chan<- ValidatedResponse) error
}

// ResponseWithSignature encapsulates a response with its signature from an operator.
// It contains the original data, a cryptographic digest, the BLS signature,
// operator identification, and the original request that triggered this response.
type ResponseWithSignature struct {
	Data        []byte
	Digest      [32]byte
	Signature   *bls.Signature
	OperatorID  [32]byte
	RequestData avsitypes.DVSRequest
}

// ValidatedResponse represents the result of a successful signature aggregation.
// It contains the aggregated data, any errors encountered, cryptographic hash,
// public keys of non-signers, group aggregate public keys, signer information,
// and various bitmap indices used for on-chain verification.
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

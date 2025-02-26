package aggregator

import (
	"github.com/0xPellNetwork/pelldvs-libs/crypto/bls"
	avsitypes "github.com/0xPellNetwork/pelldvs/avsi/types"
)

type Aggregator interface {
	// CollectResponseSignature is a method that collects the response signature from the operator
	CollectResponseSignature(response *ResponseWithSignature, result chan<- ValidatedResponse) error
}

type ResponseWithSignature struct {
	Data        []byte
	Digest      [32]byte
	Signature   *bls.Signature
	OperatorID  [32]byte
	RequestData avsitypes.DVSRequest
}

type ValidatedResponse struct {
	Data                        []byte
	Err                         error
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

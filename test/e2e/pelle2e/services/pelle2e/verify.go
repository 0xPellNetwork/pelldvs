package pelle2e

import (
	"context"
	"fmt"
	"math/big"

	"github.com/0xPellNetwork/pell-middleware-contracts/pkg/src/mocks/mockdvsservicemanager.sol"

	"github.com/0xPellNetwork/pelldvs/crypto/bls"
	coretypes "github.com/0xPellNetwork/pelldvs/rpc/core/types"
)

func (per *PellDVSE2ERunner) callVerify(eectx *E2EContext, requestResult *coretypes.ResultDvsRequest) error {
	per.logger.Info("callVerify", "eectx", eectx, "requestResult", requestResult)

	// Construct NonSignerStakesAndSignature parameters
	nonSignerPubkeysG1 := make([]mockdvsservicemanager.BN254G1Point, len(requestResult.DvsResponse.NonSignersPubkeysG1))
	for i, pubkey := range requestResult.DvsResponse.NonSignersPubkeysG1 {
		nonSignerPubkeysG1[i] = mockdvsservicemanager.BN254G1Point{
			X: new(big.Int).SetBytes(pubkey[:32]),
			Y: new(big.Int).SetBytes(pubkey[32:]),
		}
	}

	var groupApksG1 []mockdvsservicemanager.BN254G1Point
	for _, apk := range requestResult.DvsResponse.GroupApksG1 {
		tapk := bls.NewZeroG1Point()
		_ = tapk.Unmarshal(apk)
		groupApksG1 = append(groupApksG1, convertToBN254G1Point(tapk))
	}

	signersApkG2 := mockdvsservicemanager.BN254G2Point{
		X: [2]*big.Int{
			new(big.Int).SetBytes(requestResult.DvsResponse.SignersApkG2[:32]),
			new(big.Int).SetBytes(requestResult.DvsResponse.SignersApkG2[32:64]),
		},
		Y: [2]*big.Int{
			new(big.Int).SetBytes(requestResult.DvsResponse.SignersApkG2[64:96]),
			new(big.Int).SetBytes(requestResult.DvsResponse.SignersApkG2[96:]),
		},
	}

	signersAggSigG1 := mockdvsservicemanager.BN254G1Point{
		X: new(big.Int).SetBytes(requestResult.DvsResponse.SignersAggSigG1[:32]),
		Y: new(big.Int).SetBytes(requestResult.DvsResponse.SignersAggSigG1[32:]),
	}

	nonSignerStakeIndices := make([][]uint32, len(requestResult.DvsResponse.NonSignerStakeIndices))
	for i, indices := range requestResult.DvsResponse.NonSignerStakeIndices {
		nonSignerStakeIndices[i] = indices.NonSignerStakeIndice
	}

	nonSignerStakesAndSignature := mockdvsservicemanager.IBLSSignatureVerifierNonSignerStakesAndSignature{
		NonSignerPubkeys:            nonSignerPubkeysG1,
		GroupApks:                   groupApksG1,
		ApkG2:                       signersApkG2,
		Sigma:                       signersAggSigG1,
		NonSignerGroupBitmapIndices: requestResult.DvsResponse.NonSignerGroupBitmapIndices,
		GroupApkIndices:             requestResult.DvsResponse.GroupApkIndices,
		TotalStakeIndices:           requestResult.DvsResponse.TotalStakeIndices,
		NonSignerStakeIndices:       nonSignerStakeIndices,
	}

	blockNumber := uint32(requestResult.DvsRequest.Height)
	currentBlockNumber, _ := per.Client.BlockNumber(context.TODO())

	var groupNumbers []byte
	for _, groupNumber := range requestResult.DvsRequest.GroupNumbers {
		groupNumbers = append(groupNumbers, byte(groupNumber))
	}

	msgHash := eectx.KVStoreApp.GenResponseDigest()

	fmt.Println()
	fmt.Println()

	per.logger.Info("params for signature verification",
		"taskBlockNumber", blockNumber,
		"currentBlockNumber", currentBlockNumber,
		"msgHash", msgHash,
		"groupNumbers", groupNumbers,
		"nonSignerStakesAndSignature", nonSignerStakesAndSignature,
	)

	_, _, err := per.serviceManager.CheckSignatures(nil, msgHash, groupNumbers, blockNumber, nonSignerStakesAndSignature)
	if err != nil {
		per.logger.Error("Signature verification failed",
			"error", err,
			"blockNumber", blockNumber,
			"currentBlock", currentBlockNumber,
		)
		return err
	}

	per.logger.Info("✅ successfully verified aggregate signature")

	return nil
}

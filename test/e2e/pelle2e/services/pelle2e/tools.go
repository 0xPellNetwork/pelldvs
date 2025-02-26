package pelle2e

import (
	"math/big"

	"github.com/0xPellNetwork/pell-middleware-contracts/pkg/src/mocks/mockdvsservicemanager.sol"

	"github.com/0xPellNetwork/pelldvs/crypto/bls"
)

func convertToBN254G1Point(input *bls.G1Point) mockdvsservicemanager.BN254G1Point {
	output := mockdvsservicemanager.BN254G1Point{
		X: input.X.BigInt(big.NewInt(0)),
		Y: input.Y.BigInt(big.NewInt(0)),
	}
	return output
}

func convertToBN254G2Point(input *bls.G2Point) mockdvsservicemanager.BN254G2Point {
	output := mockdvsservicemanager.BN254G2Point{
		X: [2]*big.Int{input.X.A1.BigInt(big.NewInt(0)), input.X.A0.BigInt(big.NewInt(0))},
		Y: [2]*big.Int{input.Y.A1.BigInt(big.NewInt(0)), input.Y.A0.BigInt(big.NewInt(0))},
	}
	return output
}

func convertToBN254G2Point2(p []byte) mockdvsservicemanager.BN254G2Point {
	if len(p) != 128 {
		panic("invalid G2 point length")
	}
	return mockdvsservicemanager.BN254G2Point{
		X: [2]*big.Int{
			new(big.Int).SetBytes(p[32:64]), // X[1]
			new(big.Int).SetBytes(p[:32]),   // X[0]
		},
		Y: [2]*big.Int{
			new(big.Int).SetBytes(p[96:]),   // Y[1]
			new(big.Int).SetBytes(p[64:96]), // Y[0]
		},
	}
}

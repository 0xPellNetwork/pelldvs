package types

import (
	"encoding/hex"
	"log/slog"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/0xPellNetwork/pelldvs-libs/crypto/bls"
)

type OperatorID [32]byte

func (o *OperatorID) LogValue() slog.Value {
	return slog.StringValue(hex.EncodeToString(o[:]))
}

// GenOperatorIDOffChain by bls pubkey
func GenOperatorIDOffChain(pubkey *bls.G1Point) OperatorID {
	x := pubkey.X.BigInt(new(big.Int))
	y := pubkey.Y.BigInt(new(big.Int))
	return OperatorID(crypto.Keccak256Hash(append(math.U256Bytes(x), math.U256Bytes(y)...)))
}

// GenOperatorIDByAddress by eth address
func GenOperatorIDByAddress(operatorAddress common.Address) OperatorID {
	operatorID := crypto.Keccak256Hash(operatorAddress.Bytes())
	return OperatorID(operatorID[:])
}

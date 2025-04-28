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

// OperatorID uniquely identifies an operator in the distributed validation system
// Represented as a 32-byte array derived from either a BLS public key or Ethereum address
type OperatorID [32]byte

// LogValue implements slog.LogValuer interface to provide a human-readable
// string representation of the OperatorID for logging purposes
func (o *OperatorID) LogValue() slog.Value {
	return slog.StringValue(hex.EncodeToString(o[:]))
}

// OperatorIDFromBLSKey generates an OperatorID from a BLS public key
// Used for off-chain identification of operators based on their cryptographic credentials
func OperatorIDFromBLSKey(pubkey *bls.G1Point) OperatorID {
	// Convert G1Point coordinates to big integers
	x := pubkey.X.BigInt(new(big.Int))
	y := pubkey.Y.BigInt(new(big.Int))

	// Combine the X and Y coordinates and hash them to generate the ID
	return OperatorID(crypto.Keccak256Hash(append(math.U256Bytes(x), math.U256Bytes(y)...)))
}

// OperatorIDFromAddress generates an OperatorID from an Ethereum address
// Provides a way to identify operators based on their blockchain identity
func OperatorIDFromAddress(address common.Address) OperatorID {
	// Hash the Ethereum address to generate the operator ID
	operatorID := crypto.Keccak256Hash(address.Bytes())
	return OperatorID(operatorID[:])
}

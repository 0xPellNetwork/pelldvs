package bls

import (
	"bytes"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	avsiTypes "github.com/0xPellNetwork/pelldvs/avsi/types"
	avsicrypto "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/crypto"
)

const (
	KeyType = "bls"
)

var _ crypto.PubKey = PubKey{}

type PubKey struct {
	G1Point G1Point
	G2Point G2Point
}

func NewBlsPubKeyFromProto(pubkeys avsiTypes.OperatorPubkeys) *PubKey {
	return &PubKey{
		*NewZeroG1Point().Deserialize(pubkeys.G1Pubkey),
		*NewZeroG2Point().Deserialize(pubkeys.G2Pubkey),
	}
}

// Address returns the address derived from the public key
func (p PubKey) Address() crypto.Address {
	return crypto.Address(ethcrypto.Keccak256(p.G1Point.Serialize()))
}

// Bytes returns the byte representation of the public key
func (p PubKey) Bytes() []byte {
	keys := avsicrypto.OperatorPubkeys{
		G1Pubkey: p.G1Point.Serialize(),
		G2Pubkey: p.G2Point.Serialize(),
	}
	encodedBytes, _ := keys.Marshal()
	return encodedBytes
}

// VerifySignature verifies the signature of a message against the public key
func (p PubKey) VerifySignature(msg []byte, _ []byte) bool {
	if len(msg) != 32 {
		return false
	}

	var message [32]byte
	copy(message[:], msg[:32])

	signature := Signature{&p.G1Point}
	ok, err := signature.Verify(&p.G2Point, message)
	if err != nil {
		return false
	}
	return ok
}

// Equals checks whether the G 1G2 public keys are equal
func (p PubKey) Equals(other crypto.PubKey) bool {
	if otherEd, ok := other.(PubKey); ok {
		return bytes.Equal(p.Bytes(), otherEd.Bytes())
	}
	return false

}

func (pubKey PubKey) Type() string {
	return KeyType
}

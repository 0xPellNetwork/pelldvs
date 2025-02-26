package bls

import (
	"bytes"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/0xPellNetwork/pelldvs/crypto"
)

func generateTestPubKey() PubKey {
	g1Point := NewZeroG1Point()
	g2Point := NewZeroG2Point()

	return PubKey{
		G1Point: *g1Point,
		G2Point: *g2Point,
	}
}

func TestAddress(t *testing.T) {
	pubKey := generateTestPubKey()
	address := pubKey.Address()
	expectedAddress := crypto.Address(ethcrypto.Keccak256(pubKey.G1Point.Serialize()))
	if !bytes.Equal(address, expectedAddress) {
		t.Errorf("Expected address %v, got %v", expectedAddress, address)
	}
}

func TestBytes(t *testing.T) {
	pubKey := generateTestPubKey()

	expectedBytes := pubKey.Bytes()
	if !bytes.Equal(pubKey.Bytes(), expectedBytes) {
		t.Errorf("Expected bytes %v, got %v", expectedBytes, pubKey.Bytes())
	}
}

func TestEquals(t *testing.T) {
	pubKey1 := generateTestPubKey()
	pubKey2 := generateTestPubKey()

	if !pubKey1.Equals(pubKey2) {
		t.Errorf("Expected pubKey1 to equal pubKey2")
	}
}

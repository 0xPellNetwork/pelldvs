package types

import (
	"errors"

	"github.com/0xPellNetwork/pelldvs/crypto"
	"github.com/0xPellNetwork/pelldvs/crypto/bls"
	"github.com/0xPellNetwork/pelldvs/crypto/ed25519"
)

// PrivValidator defines the functionality of a local PellDVS validator
// that signs votes and proposals, and never double signs.
type PrivValidator interface {
	GetPubKey() (*bls.G1Point, error)
	SignBytes(bytes []byte) (*bls.Signature, error)
}

//----------------------------------------
// MockPV

// MockPV implements PrivValidator without any safety or persistence.
// Only use it for testing.
type MockPV struct {
	PrivKey              crypto.PrivKey
	breakProposalSigning bool
	breakVoteSigning     bool
}

func NewMockPV() MockPV {
	return MockPV{ed25519.GenPrivKey(), false, false}
}

// NewMockPVWithParams allows one to create a MockPV instance, but with finer
// grained control over the operation of the mock validator. This is useful for
// mocking test failures.
func NewMockPVWithParams(privKey crypto.PrivKey, breakProposalSigning, breakVoteSigning bool) MockPV {
	return MockPV{privKey, breakProposalSigning, breakVoteSigning}
}

// Implements PrivValidator.
func (pv MockPV) GetPubKey() (crypto.PubKey, error) {
	return pv.PrivKey.PubKey(), nil
}

type ErroringMockPV struct {
	MockPV
}

var ErroringMockPVErr = errors.New("erroringMockPV always returns an error")

// NewErroringMockPV returns a MockPV that fails on each signing request. Again, for testing only.

func NewErroringMockPV() *ErroringMockPV {
	return &ErroringMockPV{MockPV{ed25519.GenPrivKey(), false, false}}
}

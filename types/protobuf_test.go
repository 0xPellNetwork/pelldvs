package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/pelldvs/crypto"
	"github.com/0xPellNetwork/pelldvs/crypto/ed25519"
	cryptoenc "github.com/0xPellNetwork/pelldvs/crypto/encoding"
)

func TestABCIPubKey(t *testing.T) {
	pkEd := ed25519.GenPrivKey().PubKey()
	err := testABCIPubKey(t, pkEd)
	assert.NoError(t, err)
}

func testABCIPubKey(t *testing.T, pk crypto.PubKey) error {
	abciPubKey, err := cryptoenc.PubKeyToProto(pk)
	require.NoError(t, err)
	pk2, err := cryptoenc.PubKeyFromProto(abciPubKey)
	require.NoError(t, err)
	require.Equal(t, pk, pk2)
	return nil
}

package privval

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenLoadValidator(t *testing.T) {
	privVal, tempKeyFileName, _ := newTestFilePV(t)

	privVal.Save()

	privVal = LoadFilePV(tempKeyFileName)
	t.Log(privVal.String())
}

func TestLoadOrGenValidator(t *testing.T) {
	tempKeyFile, err := os.CreateTemp("", "priv_validator_key_")
	require.Nil(t, err)

	tempKeyFilePath := tempKeyFile.Name()
	if err := os.Remove(tempKeyFilePath); err != nil {
		t.Error(err)
	}

	privVal, err := LoadOrGenFilePV(tempKeyFilePath)
	require.NoError(t, err)

	t.Log(privVal.String())
}

func newTestFilePV(t *testing.T) (*FilePV, string, string) {
	tempKeyFile, err := os.CreateTemp(t.TempDir(), "priv_validator_key_")
	require.NoError(t, err)
	tempStateFile, err := os.CreateTemp(t.TempDir(), "priv_validator_state_")
	require.NoError(t, err)

	privVal, err := GenFilePV(tempKeyFile.Name())
	require.NoError(t, err)

	return privVal, tempKeyFile.Name(), tempStateFile.Name()
}

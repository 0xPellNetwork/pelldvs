package config

// package config_test

// import (
// 	"os"
// 	"path/filepath"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"

// 	cmfcfg "github.com/0xPellNetwork/pelldvs/config"

// 	pellcfg "github.com/0xPellNetwork/pelldvs/config"

// 	"github.com/0xPellNetwork/pelldvs/internal/test"
// )

// func ensureFiles(t *testing.T, rootDir string, files ...string) {
// 	for _, f := range files {
// 		p := filepath.Join(rootDir, f)
// 		_, err := os.Stat(p)
// 		assert.NoError(t, err, p)
// 	}
// }

// func TestEnsureRoot(t *testing.T) {
// 	require := require.New(t)

// 	// setup temp dir for test
// 	tmpDir, err := os.MkdirTemp("", "config-test")
// 	require.Nil(err)
// 	defer os.RemoveAll(tmpDir)

// 	// create root dir
// 	pellcfg.EnsureRoot(tmpDir)

// 	// make sure config is set properly
// 	data, err := os.ReadFile(filepath.Join(tmpDir, cmfcfg.DefaultConfigDir, cmfcfg.DefaultConfigFileName))
// 	require.Nil(err)

// 	assertValidConfig(t, string(data))

// 	ensureFiles(t, tmpDir, "data")
// }

// func TestEnsureTestRoot(t *testing.T) {
// 	require := require.New(t)

// 	// create root dir
// 	cfg := test.ResetTestRoot("ensureTestRoot")
// 	defer os.RemoveAll(cfg.RootDir)
// 	rootDir := cfg.RootDir
// 	pellcfg.WriteConfigFile(filepath.Join(rootDir, cmfcfg.DefaultConfigDir, cmfcfg.DefaultConfigFileName), cfg, nil)

// 	// make sure config is set properly
// 	data, err := os.ReadFile(filepath.Join(rootDir, cmfcfg.DefaultConfigDir, cmfcfg.DefaultConfigFileName))
// 	require.Nil(err)

// 	assertValidConfig(t, string(data))

// 	// TODO: make sure the cfg returned and testconfig are the same!
// 	baseConfig := cmfcfg.DefaultBaseConfig()
// 	ensureFiles(t, rootDir, cmfcfg.DefaultDataDir, baseConfig.Genesis, baseConfig.PrivValidatorKey, baseConfig.PrivValidatorState)
// }

// func assertValidConfig(t *testing.T, configFile string) {
// 	t.Helper()
// 	// list of words we expect in the config
// 	var elems = []string{
// 		"moniker",
// 		"seeds",
// 		"proxy_app",
// 		"create_empty_blocks",
// 		"peer",
// 		"timeout",
// 		"broadcast",
// 		"send",
// 		"addr",
// 		"wal",
// 		"propose",
// 		"max",
// 		"genesis",
// 		"delegation_manager_address",
// 		"registry_router_address",
// 	}
// 	for _, e := range elems {
// 		assert.Contains(t, configFile, e)
// 	}
// }

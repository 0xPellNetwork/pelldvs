package debug

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	cfg "github.com/0xPellNetwork/pelldvs/config"
	rpchttp "github.com/0xPellNetwork/pelldvs/rpc/client/http"
)

// dumpNetInfo gets network information state dump from the PellDVS RPC and
// writes it to file. It returns an error upon failure.
func dumpNetInfo(rpc *rpchttp.HTTP, dir, filename string) error {
	netInfo, err := rpc.NetInfo(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get node network information: %w", err)
	}

	return writeStateJSONToFile(netInfo, dir, filename)
}

// copyWAL copies the PellDVS node's WAL file. It returns an error if the
// WAL file cannot be read or copied.
func copyWAL(conf *cfg.Config, dir string) error {

	return nil
	// walPath := conf.Consensus.WalFile()
	// walFile := filepath.Base(walPath)

	// return copyFile(walPath, filepath.Join(dir, walFile))
}

// copyConfig copies the PellDVS node's config file. It returns an error if
// the config file cannot be read or copied.
func copyConfig(home, dir string) error {
	configFile := "config.toml"
	configPath := filepath.Join(home, "config", configFile)

	return copyFile(configPath, filepath.Join(dir, configFile))
}

func dumpProfile(dir, addr, profile string, debug int) error {
	endpoint := fmt.Sprintf("%s/debug/pprof/%s?debug=%d", addr, profile, debug)

	//nolint:gosec,nolintlint
	resp, err := http.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to query for %s profile: %w", profile, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read %s profile response body: %w", profile, err)
	}

	return os.WriteFile(path.Join(dir, fmt.Sprintf("%s.out", profile)), body, 0o600)
}

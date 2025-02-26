package app

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	avsi "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/version"
)

const (
	appVersion                 = 1
	voteExtensionKey    string = "extensionSum"
	voteExtensionMaxVal int64  = 128
	prefixReservedKey   string = "reservedTxKey_"
	suffixChainID       string = "ChainID"
	suffixVoteExtHeight string = "VoteExtensionsHeight"
	suffixInitialHeight string = "InitialHeight"
)

// Application is an ABCI application for use by end-to-end tests. It is a
// simple key/value store for strings, storing data in memory and persisting
// to disk as JSON, taking state sync snapshots if requested.
type Application struct {
	avsi.BaseApplication
	logger        log.Logger
	state         *State
	cfg           *Config
	restoreChunks [][]byte
}

// Config allows for the setting of high level parameters for running the e2e Application
// KeyType and ValidatorUpdates must be the same for all nodes running the same application.
type Config struct {
	// The directory with which state.json will be persisted in. Usually $HOME/.pelldvs/data
	Dir string `toml:"dir"`

	// SnapshotInterval specifies the height interval at which the application
	// will take state sync snapshots. Defaults to 0 (disabled).
	SnapshotInterval uint64 `toml:"snapshot_interval"`

	// RetainBlocks specifies the number of recent blocks to retain. Defaults to
	// 0, which retains all blocks. Must be greater that PersistInterval,
	// SnapshotInterval and EvidenceAgeHeight.
	RetainBlocks uint64 `toml:"retain_blocks"`

	// KeyType sets the curve that will be used by validators.
	// Options are ed25519 & secp256k1
	KeyType string `toml:"key_type"`

	// PersistInterval specifies the height interval at which the application
	// will persist state to disk. Defaults to 1 (every height), setting this to
	// 0 disables state persistence.
	PersistInterval uint64 `toml:"persist_interval"`

	// ValidatorUpdates is a map of heights to validator names and their power,
	// and will be returned by the ABCI application. For example, the following
	// changes the power of validator01 and validator02 at height 1000:
	//
	// [validator_update.1000]
	// validator01 = 20
	// validator02 = 10
	//
	// Specifying height 0 returns the validator update during InitChain. The
	// application returns the validator updates as-is, i.e. removing a
	// validator must be done by returning it with power 0, and any validators
	// not specified are not changed.
	//
	// height <-> pubkey <-> voting power
	ValidatorUpdates map[string]map[string]uint8 `toml:"validator_update"`

	// Add artificial delays to each of the main ABCI calls to mimic computation time
	// of the application
	PrepareProposalDelay time.Duration `toml:"prepare_proposal_delay"`
	ProcessProposalDelay time.Duration `toml:"process_proposal_delay"`
	CheckTxDelay         time.Duration `toml:"check_tx_delay"`
	FinalizeBlockDelay   time.Duration `toml:"finalize_block_delay"`
	VoteExtensionDelay   time.Duration `toml:"vote_extension_delay"`

	// VoteExtensionsEnableHeight configures the first height during which
	// the chain will use and require vote extension data to be present
	// in precommit messages.
	VoteExtensionsEnableHeight int64 `toml:"vote_extensions_enable_height"`

	// VoteExtensionsUpdateHeight configures the height at which consensus
	// param VoteExtensionsEnableHeight will be set.
	// -1 denotes it is set at genesis.
	// 0 denotes it is set at InitChain.
	VoteExtensionsUpdateHeight int64 `toml:"vote_extensions_update_height"`
}

func DefaultConfig(dir string) *Config {
	return &Config{
		PersistInterval:  1,
		SnapshotInterval: 100,
		Dir:              dir,
	}
}

// NewApplication creates the application.
func NewApplication(cfg *Config) (*Application, error) {
	state, err := NewState(cfg.Dir, cfg.PersistInterval)
	if err != nil {
		return nil, err
	}
	return &Application{
		logger: log.NewLogger(os.Stdout),
		state:  state,
		cfg:    cfg,
	}, nil
}

// Info implements ABCI.
func (app *Application) Info(context.Context, *avsi.RequestInfo) (*avsi.ResponseInfo, error) {
	height, hash := app.state.Info()
	return &avsi.ResponseInfo{
		Version:          version.ABCIVersion,
		AppVersion:       appVersion,
		LastBlockHeight:  int64(height),
		LastBlockAppHash: hash,
	}, nil
}

// Query implements ABCI.
func (app *Application) Query(_ context.Context, req *avsi.RequestQuery) (*avsi.ResponseQuery, error) {
	value, height := app.state.Query(string(req.Data))
	return &avsi.ResponseQuery{
		Height: int64(height),
		Key:    req.Data,
		Value:  []byte(value),
	}, nil
}

func (app *Application) getAppHeight() int64 {
	initialHeightStr, height := app.state.Query(prefixReservedKey + suffixInitialHeight)
	if len(initialHeightStr) == 0 {
		panic("initial height not set in database")
	}
	initialHeight, err := strconv.ParseInt(initialHeightStr, 10, 64)
	if err != nil {
		panic(fmt.Errorf("malformed initial height %q in database", initialHeightStr))
	}

	appHeight := int64(height)
	if appHeight == 0 {
		appHeight = initialHeight - 1
	}
	return appHeight + 1
}

func (app *Application) checkHeightAndExtensions(isPrepareProcessProposal bool, height int64, callsite string) (int64, bool) {
	appHeight := app.getAppHeight()
	if height != appHeight {
		panic(fmt.Errorf(
			"got unexpected height in %s request; expected %d, actual %d",
			callsite, appHeight, height,
		))
	}

	voteExtHeightStr := app.state.Get(prefixReservedKey + suffixVoteExtHeight)
	if len(voteExtHeightStr) == 0 {
		panic("vote extension height not set in database")
	}
	voteExtHeight, err := strconv.ParseInt(voteExtHeightStr, 10, 64)
	if err != nil {
		panic(fmt.Errorf("malformed vote extension height %q in database", voteExtHeightStr))
	}
	currentHeight := appHeight
	if isPrepareProcessProposal {
		currentHeight-- // at exactly voteExtHeight, PrepareProposal still has no extensions, see RFC100
	}

	return appHeight, voteExtHeight != 0 && currentHeight >= voteExtHeight
}

// parseTx parses a tx in 'key=value' format into a key and value.
func parseTx(tx []byte) (string, string, error) {
	parts := bytes.Split(tx, []byte("="))
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid tx format: %q", string(tx))
	}
	if len(parts[0]) == 0 {
		return "", "", errors.New("key cannot be empty")
	}
	return string(parts[0]), string(parts[1]), nil
}

// parseVoteExtension attempts to parse the given extension data into a positive
// integer value.
func parseVoteExtension(ext []byte) (int64, error) {
	num, errVal := binary.Varint(ext)
	if errVal == 0 {
		return 0, errors.New("vote extension is too small to parse")
	}
	if errVal < 0 {
		return 0, errors.New("vote extension value is too large")
	}
	if num >= voteExtensionMaxVal {
		return 0, fmt.Errorf("vote extension value must be smaller than %d (was %d)", voteExtensionMaxVal, num)
	}
	return num, nil
}

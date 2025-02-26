package rpctest

import (
	"fmt"
	"os"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	avsi "github.com/0xPellNetwork/pelldvs/avsi/types"
	cfg "github.com/0xPellNetwork/pelldvs/config"
	nm "github.com/0xPellNetwork/pelldvs/node"
	"github.com/0xPellNetwork/pelldvs/p2p"
	"github.com/0xPellNetwork/pelldvs/privval"
	"github.com/0xPellNetwork/pelldvs/proxy"
)

// Options helps with specifying some parameters for our RPC testing for greater
// control.
type Options struct {
	suppressStdout  bool
	recreateConfig  bool
	maxReqBatchSize int
}

var (
	globalConfig   *cfg.Config
	defaultOptions = Options{
		suppressStdout: false,
		recreateConfig: false,
	}
)

// GetConfig returns a config for the test cases as a singleton
func GetConfig(forceCreate ...bool) *cfg.Config {
	return globalConfig
}

// StartTendermint starts a test PellDVS server in a go routine and returns when it is initialized
func StartTendermint(app avsi.Application, opts ...func(*Options)) *nm.Node {
	nodeOpts := defaultOptions
	for _, opt := range opts {
		opt(&nodeOpts)
	}
	node := NewTendermint(app, &nodeOpts)
	err := node.Start()
	if err != nil {
		panic(err)
	}

	if !nodeOpts.suppressStdout {
		fmt.Println("PellDVS running!")
	}

	return node
}

// StopTendermint stops a test PellDVS server, waits until it's stopped and
// cleans up test/config files.
func StopTendermint(node *nm.Node) {
	if err := node.Stop(); err != nil {
		node.Logger.Error("Error when trying to stop node", "err", err)
	}
	node.Wait()
	os.RemoveAll(node.Config().RootDir)
}

// NewTendermint creates a new PellDVS server and sleeps forever
func NewTendermint(app avsi.Application, opts *Options) *nm.Node {
	// Create & start node
	config := GetConfig(opts.recreateConfig)
	var logger log.Logger
	if opts.suppressStdout {
		logger = log.NewNopLogger()
	} else {
		logger = log.NewLogger(os.Stdout)
	}
	if opts.maxReqBatchSize > 0 {
		config.RPC.MaxRequestBatchSize = opts.maxReqBatchSize
	}
	pvKeyFile := config.PrivValidatorKeyFile()
	pvKeyStateFile := config.PrivValidatorStateFile()
	pv := privval.LoadOrGenFilePV(pvKeyFile, pvKeyStateFile)
	papp := proxy.NewLocalClientCreator(app)
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		panic(err)
	}
	node, err := nm.NewNode(config, pv, nodeKey, papp,
		cfg.DefaultDBProvider,
		nil,
		nm.DefaultMetricsProvider(config.Instrumentation),
		logger)
	if err != nil {
		panic(err)
	}
	return node
}

// SuppressStdout is an option that tries to make sure the RPC test PellDVS
// node doesn't log anything to stdout.
func SuppressStdout(o *Options) {
	o.suppressStdout = true
}

// RecreateConfig instructs the RPC test to recreate the configuration each
// time, instead of treating it as a global singleton.
func RecreateConfig(o *Options) {
	o.recreateConfig = true
}

// MaxReqBatchSize is an option to limit the maximum number of requests per batch.
func MaxReqBatchSize(o *Options) {
	o.maxReqBatchSize = 2
}

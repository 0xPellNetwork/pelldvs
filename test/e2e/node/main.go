package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/avsi/server"
	"github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/node"
	"github.com/0xPellNetwork/pelldvs/p2p"
	"github.com/0xPellNetwork/pelldvs/privval"
	"github.com/0xPellNetwork/pelldvs/proxy"
	"github.com/0xPellNetwork/pelldvs/test/e2e/app"
	e2e "github.com/0xPellNetwork/pelldvs/test/e2e/pkg"
)

var logger = log.NewLogger(os.Stdout)

// main is the binary entrypoint.
func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %v <configfile>", os.Args[0])
		return
	}
	configFile := ""
	if len(os.Args) == 2 {
		configFile = os.Args[1]
	}

	if err := run(configFile); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

// run runs the application - basically like main() with error handling.
func run(configFile string) error {
	cfg, err := LoadConfig(configFile)
	if err != nil {
		return err
	}

	// Start app server.
	switch cfg.Protocol {
	case "socket", "grpc":
		err = startApp(cfg)
	default:
		err = fmt.Errorf("invalid protocol %q", cfg.Protocol)
	}
	if err != nil {
		return err
	}

	// Apparently there's no way to wait for the server, so we just sleep
	for {
		time.Sleep(1 * time.Hour)
	}
}

// startApp starts the application server, listening for connections from PellDVS.
func startApp(cfg *Config) error {
	app, err := app.NewApplication(cfg.App())
	if err != nil {
		return err
	}
	server, err := server.NewServer(cfg.Listen, cfg.Protocol, app)
	if err != nil {
		return err
	}
	err = server.Start()
	if err != nil {
		return err
	}
	logger.Info("start app", "msg", log.NewLazySprintf("Server listening on %v (%v protocol)", cfg.Listen, cfg.Protocol))
	return nil
}

// startNode starts a PellDVS node running the application directly. It assumes the PellDVS
// configuration is in $PELLDVSHOME/config/pelldvs.toml.
//
// FIXME There is no way to simply load the configuration from a file, so we need to pull in Viper.
func startNode(cfg *Config) error {
	app, err := app.NewApplication(cfg.App())
	if err != nil {
		return err
	}

	cmtcfg, nodeLogger, nodeKey, err := setupNode()
	if err != nil {
		return fmt.Errorf("failed to setup config: %w", err)
	}

	var clientCreator proxy.ClientCreator
	if cfg.Protocol == string(e2e.ProtocolBuiltinConnSync) {
		clientCreator = proxy.NewConnSyncLocalClientCreator(app)
		nodeLogger.Info("Using connection-synchronized local client creator")
	} else {
		clientCreator = proxy.NewLocalClientCreator(app)
		nodeLogger.Info("Using default (synchronized) local client creator")
	}

	pv, err := privval.LoadOrGenFilePV(cmtcfg.PrivValidatorKeyFile())
	if err != nil {
		return err
	}
	n, err := node.NewNode(cmtcfg,
		pv,
		nodeKey,
		clientCreator,
		config.DefaultDBProvider,
		nil,
		nil,
		node.DefaultMetricsProvider(cmtcfg.Instrumentation),
		nodeLogger,
	)
	if err != nil {
		return err
	}
	return n.Start()
}

func setupNode() (*config.Config, log.Logger, *p2p.NodeKey, error) {
	var cmtcfg *config.Config

	home := os.Getenv("PELLDVSHOME")
	if home == "" {
		return nil, nil, nil, errors.New("PELLDVSHOME not set")
	}

	viper.AddConfigPath(filepath.Join(home, "config"))
	viper.SetConfigName("config")

	if err := viper.ReadInConfig(); err != nil {
		return nil, nil, nil, err
	}

	cmtcfg = config.DefaultConfig()

	if err := viper.Unmarshal(cmtcfg); err != nil {
		return nil, nil, nil, err
	}

	cmtcfg.SetRoot(home)

	if err := cmtcfg.ValidateBasic(); err != nil {
		return nil, nil, nil, fmt.Errorf("error in config file: %w", err)
	}

	if cmtcfg.LogFormat == config.LogFormatJSON {
		logger = log.NewLogger(os.Stdout)
	}

	nodeLogger := logger.With("module", "main")

	nodeKey, err := p2p.LoadOrGenNodeKey(cmtcfg.NodeKeyFile())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load or gen node key %s: %w", cmtcfg.NodeKeyFile(), err)
	}

	return cmtcfg, nodeLogger, nodeKey, nil
}

// rpcEndpoints takes a list of persistent peers and splits them into a list of rpc endpoints
// using 26657 as the port number
func rpcEndpoints(peers string) []string {
	arr := strings.Split(peers, ",")
	endpoints := make([]string, len(arr))
	for i, v := range arr {
		urlString := strings.SplitAfter(v, "@")[1]
		hostName := strings.Split(urlString, ":26656")[0]
		// use RPC port instead
		port := 26657
		rpcEndpoint := "http://" + hostName + ":" + fmt.Sprint(port)
		endpoints[i] = rpcEndpoint
	}
	return endpoints
}

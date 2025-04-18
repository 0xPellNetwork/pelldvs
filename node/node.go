package node

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof" //nolint: gosec
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	"github.com/0xPellNetwork/pelldvs/aggregator"
	cfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/libs/service"
	"github.com/0xPellNetwork/pelldvs/p2p"
	"github.com/0xPellNetwork/pelldvs/p2p/pex"
	"github.com/0xPellNetwork/pelldvs/proxy"
	rpccore "github.com/0xPellNetwork/pelldvs/rpc/core"
	grpccore "github.com/0xPellNetwork/pelldvs/rpc/grpc"
	rpcserver "github.com/0xPellNetwork/pelldvs/rpc/jsonrpc/server"
	"github.com/0xPellNetwork/pelldvs/security"
	"github.com/0xPellNetwork/pelldvs/state/requestindex"
	"github.com/0xPellNetwork/pelldvs/state/requestindex/null"
	"github.com/0xPellNetwork/pelldvs/types"
	"github.com/0xPellNetwork/pelldvs/version"
)

// Node is the highest level interface to a full PellDVS node.
// It includes all configuration information and running services.
type Node struct {
	service.BaseService

	// config
	config *cfg.Config

	privValidator types.PrivValidator // local node's validator key

	// network
	transport   *p2p.MultiplexTransport
	sw          *p2p.Switch  // p2p connections
	addrBook    pex.AddrBook // known peers
	nodeInfo    p2p.NodeInfo
	nodeKey     *p2p.NodeKey // our node privkey
	isListening bool

	// services
	proxyApp          proxy.AppConns // connection to the application
	dvsReactor        security.DVSReactor
	aggregatorReactor *security.AggregatorReactor

	rpcListeners []net.Listener // rpc servers
	pexReactor   *pex.Reactor   // for exchanging peer addresses

	prometheusSrv *http.Server
	pprofSrv      *http.Server

	dvsRequestIndexer requestindex.DvsRequestIndexer
}

// Option sets a parameter for the node.
type Option func(*Node)

// CustomReactors allows you to add custom reactors (name -> p2p.Reactor) to
// the node's Switch.
//
// WARNING: using any name from the below list of the existing reactors will
// result in replacing it with the custom one.
//
//   - MEMPOOL
//   - BLOCKSYNC
//   - CONSENSUS
//   - EVIDENCE
//   - PEX
//   - STATESYNC
func CustomReactors(reactors map[string]p2p.Reactor) Option {
	return func(n *Node) {
		for name, reactor := range reactors {
			if existingReactor := n.sw.Reactor(name); existingReactor != nil {
				n.sw.Logger.Info("Replacing existing reactor with a custom one",
					"name", name, "existing", existingReactor, "custom", reactor)
				n.sw.RemoveReactor(name, existingReactor)
			}
			n.sw.AddReactor(name, reactor)
			// register the new channels to the nodeInfo
			// NOTE: This is a bit messy now with the type casting but is
			// cleaned up in the following version when NodeInfo is changed from
			// and interface to a concrete type
			if ni, ok := n.nodeInfo.(p2p.DefaultNodeInfo); ok {
				for _, chDesc := range reactor.GetChannels() {
					if !ni.HasChannel(chDesc.ID) {
						ni.Channels = append(ni.Channels, chDesc.ID)
						n.transport.AddChannel(chDesc.ID)
					}
				}
				n.nodeInfo = ni
			} else {
				n.Logger.Error("Node info is not of type DefaultNodeInfo. Custom reactor channels can not be added.")
			}
		}
	}
}

//------------------------------------------------------------------------------

// NewNode returns a new, ready to go, PellDVS Node.
func NewNode(config *cfg.Config,
	privValidator types.PrivValidator,
	nodeKey *p2p.NodeKey,
	clientCreator proxy.ClientCreator,
	dbProvider cfg.DBProvider,
	aggregator aggregator.Aggregator,
	metricsProvider MetricsProvider,
	logger log.Logger,
	options ...Option,
) (*Node, error) {
	return NewNodeWithContext(context.TODO(), config, privValidator,
		nodeKey, clientCreator, dbProvider, aggregator,
		metricsProvider, logger, options...)
}

// NewNodeWithContext is cancellable version of NewNode.
func NewNodeWithContext(ctx context.Context,
	config *cfg.Config,
	privValidator types.PrivValidator,
	nodeKey *p2p.NodeKey,
	clientCreator proxy.ClientCreator,
	dbProvider cfg.DBProvider,
	aggregator aggregator.Aggregator,
	metricsProvider MetricsProvider,
	logger log.Logger,
	options ...Option,
) (*Node, error) {

	// TODO: add service id from config
	p2pMetrics, avsiMetrics := metricsProvider("id")

	// Create the proxyApp and establish connections to the AVSI app (consensus, mempool, query).
	proxyApp, err := createAndStartProxyAppConns(clientCreator, logger, avsiMetrics)
	if err != nil {
		return nil, err
	}

	// If an address is provided, listen on the socket for a connection from an
	// external signing process.
	if config.PrivValidatorListenAddr != "" {
		// TODO: add service id from config
		// FIXME: we should start services inside OnStart
		privValidator, err = createAndStartPrivValidatorSocketClient(config.PrivValidatorListenAddr, "ID", logger)
		if err != nil {
			return nil, fmt.Errorf("error with private validator socket client: %w", err)
		}
	}
	_ = privValidator

	// EventBus and IndexerService must be started before the handshake because
	// we might need to index the txs of the replayed block as this might not have happened
	// when the node stopped last time (i.e. the node stopped after it saved the block
	// but before it indexed the txs)
	// eventBus, err := createAndStartEventBus(logger)
	// if err != nil {
	// 	return nil, err
	// }

	// If an address is provided, listen on the socket for a connection from an
	// external signing process.

	dvsRequestIndexer, err := createDvsRequestIndexer(config, dbProvider, logger)

	if err != nil {
		return nil, err
	}

	nodeInfo, err := makeNodeInfo(config, nodeKey, dvsRequestIndexer)
	if err != nil {
		return nil, err
	}

	transport, peerFilters := createTransport(config, nodeInfo, nodeKey, proxyApp)

	p2pLogger := logger.With("module", "p2p")

	sw := createSwitch(
		config, transport, p2pMetrics, peerFilters, nodeInfo, nodeKey, p2pLogger,
	)

	err = sw.AddPersistentPeers(splitAndTrimEmpty(config.P2P.PersistentPeers, ",", " "))
	if err != nil {
		return nil, fmt.Errorf("could not add peers from persistent_peers field: %w", err)
	}

	err = sw.AddUnconditionalPeerIDs(splitAndTrimEmpty(config.P2P.UnconditionalPeerIDs, ",", " "))
	if err != nil {
		return nil, fmt.Errorf("could not add peer ids from unconditional_peer_ids field: %w", err)
	}

	addrBook, err := createAddrBookAndSetOnSwitch(config, sw, p2pLogger, nodeKey)
	if err != nil {
		return nil, fmt.Errorf("could not create addrbook: %w", err)
	}

	db, err := cfg.DefaultDBProvider(&cfg.DBContext{
		ID:     "indexer",
		Config: config,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init db: %v", err)
	}
	// Optionally, start the pex reactor
	//
	// TODO:
	//
	// We need to set Seeds and PersistentPeers on the switch,
	// since it needs to be able to use these (and their DNS names)
	// even if the PEX is off. We can include the DNS name in the NetAddress,
	// but it would still be nice to have a clear list of the current "PersistentPeers"
	// somewhere that we can return with net_info.
	//
	// If PEX is on, it should handle dialing the seeds. Otherwise the switch does it.
	// Note we currently use the addrBook regardless at least for AddOurAddress
	var pexReactor *pex.Reactor
	if config.P2P.PexReactor {
		pexReactor = createPEXReactorAndAddToSwitch(addrBook, config, sw, logger)
	}

	// Add private IDs to addrbook to block those peers being added
	addrBook.AddPrivateIDs(splitAndTrimEmpty(config.P2P.PrivatePeerIDs, ",", " "))

	// Initialize the DVS request store
	storeDir := config.RootDir + "/data/security_store"
	dvsReqStore, err := security.NewPersistentStore(storeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create DVS request store: %v", err)
	}

	dvsState, err := security.NewDVSState(config.Pell, dvsReqStore, storeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create DVS state: %v", err)
	}

	// Create the event manager
	eventManager := security.NewEventManager(logger)

	// Create the DVS and Aggregator reactors
	dvsReactor, err := security.CreateDVSReactor(*config.Pell, proxyApp, dvsRequestIndexer, db, dvsState, logger, eventManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create dvsReactor: %w", err)
	}
	aggregatorReactor := security.CreateAggregatorReactor(aggregator, dvsRequestIndexer, privValidator, dvsState, logger, eventManager)

	eventManager.SetDVSReactor(&dvsReactor)
	eventManager.SetAggregatorReactor(aggregatorReactor)
	eventManager.StartListening()

	node := &Node{
		config:     config,
		transport:  transport,
		sw:         sw,
		addrBook:   addrBook,
		nodeInfo:   nodeInfo,
		nodeKey:    nodeKey,
		proxyApp:   proxyApp,
		pexReactor: pexReactor,

		dvsRequestIndexer: dvsRequestIndexer,
		dvsReactor:        dvsReactor,
		aggregatorReactor: aggregatorReactor,
	}
	node.BaseService = *service.NewBaseService(logger, "Node", node)

	for _, option := range options {
		option(node)
	}

	return node, nil
}

// OnStart starts the Node. It implements service.Service.
func (n *Node) OnStart() error {

	// run pprof server if it is enabled
	if n.config.RPC.IsPprofEnabled() {
		n.pprofSrv = n.startPprofServer()
	}

	// begin prometheus metrics gathering if it is enabled
	if n.config.Instrumentation.IsPrometheusEnabled() {
		n.prometheusSrv = n.startPrometheusServer()
	}

	// Start the RPC server before the P2P server
	// so we can eg. receive txs for the first block
	if n.config.RPC.ListenAddress != "" {
		listeners, err := n.startRPC()
		if err != nil {
			return err
		}
		n.rpcListeners = listeners
	}

	// Start the transport.
	addr, err := p2p.NewNetAddressString(p2p.IDAddressString(n.nodeKey.ID(), n.config.P2P.ListenAddress))
	if err != nil {
		return err
	}
	if err := n.transport.Listen(*addr); err != nil {
		return err
	}

	n.isListening = true

	// Start the switch (the P2P server).
	err = n.sw.Start()
	if err != nil {
		return err
	}

	// Always connect to persistent peers
	err = n.sw.DialPeersAsync(splitAndTrimEmpty(n.config.P2P.PersistentPeers, ",", " "))
	if err != nil {
		return fmt.Errorf("could not dial peers from persistent_peers field: %w", err)
	}

	return nil
}

// OnStop stops the Node. It implements service.Service.
func (n *Node) OnStop() {
	n.BaseService.OnStop()

	n.Logger.Info("Stopping Node")

	// // first stop the non-reactor services
	// if err := n.eventBus.Stop(); err != nil {
	// 	n.Logger.Error("Error closing eventBus", "err", err)
	// }

	if pvsc, ok := n.privValidator.(service.Service); ok {
		if err := pvsc.Stop(); err != nil {
			n.Logger.Error("Error closing private validator", "err", err)
		}
	}

	// now stop the reactors
	if err := n.sw.Stop(); err != nil {
		n.Logger.Error("Error closing switch", "err", err)
	}

	if err := n.transport.Close(); err != nil {
		n.Logger.Error("Error closing transport", "err", err)
	}

	n.isListening = false

	// finally stop the listeners / external services
	for _, l := range n.rpcListeners {
		n.Logger.Info("Closing rpc listener", "listener", l)
		if err := l.Close(); err != nil {
			n.Logger.Error("Error closing listener", "listener", l, "err", err)
		}
	}

	if n.prometheusSrv != nil {
		if err := n.prometheusSrv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			n.Logger.Error("Prometheus HTTP server Shutdown", "err", err)
		}
	}
	if n.pprofSrv != nil {
		if err := n.pprofSrv.Shutdown(context.Background()); err != nil {
			n.Logger.Error("Pprof HTTP server Shutdown", "err", err)
		}
	}
}

// ConfigureRPC makes sure RPC has all the objects it needs to operate.
func (n *Node) ConfigureRPC() (*rpccore.Environment, error) {

	rpcCoreEnv := rpccore.Environment{
		ProxyAppQuery: n.proxyApp.Query(),
		DVSReactor:    n.dvsReactor,
		P2PPeers:      n.sw,
		P2PTransport:  n,

		DvsRequestIndexer: n.dvsRequestIndexer,

		Logger: n.Logger.With("module", "rpc"),

		Config: *n.config.RPC,
	}
	return &rpcCoreEnv, nil
}

func (n *Node) startRPC() ([]net.Listener, error) {
	env, err := n.ConfigureRPC()
	if err != nil {
		return nil, err
	}

	listenAddrs := splitAndTrimEmpty(n.config.RPC.ListenAddress, ",", " ")
	routes := env.GetRoutes()

	if n.config.RPC.Unsafe {
		env.AddUnsafeRoutes(routes)
	}

	config := rpcserver.DefaultConfig()
	config.MaxRequestBatchSize = n.config.RPC.MaxRequestBatchSize
	config.MaxBodyBytes = n.config.RPC.MaxBodyBytes
	config.MaxHeaderBytes = n.config.RPC.MaxHeaderBytes
	config.MaxOpenConnections = n.config.RPC.MaxOpenConnections
	// If necessary adjust global WriteTimeout to ensure it's greater than
	// TimeoutBroadcastTxCommit.
	// See https://github.com/tendermint/tendermint/issues/3435
	if config.WriteTimeout <= n.config.RPC.TimeoutBroadcastTxCommit {
		config.WriteTimeout = n.config.RPC.TimeoutBroadcastTxCommit + 1*time.Second
	}

	// we may expose the rpc over both a unix and tcp socket
	listeners := make([]net.Listener, len(listenAddrs))
	for i, listenAddr := range listenAddrs {
		mux := http.NewServeMux()
		rpcLogger := n.Logger.With("module", "rpc-server")
		wmLogger := rpcLogger.With("protocol", "websocket")
		wm := rpcserver.NewWebsocketManager(routes,
			rpcserver.OnDisconnect(func(remoteAddr string) {
				// err := n.eventBus.UnsubscribeAll(context.Background(), remoteAddr)
				// if err != nil && err != cmtpubsub.ErrSubscriptionNotFound {
				// 	wmLogger.Error("Failed to unsubscribe addr from events", "addr", remoteAddr, "err", err)
				// }
			}),
			rpcserver.ReadLimit(config.MaxBodyBytes),
			rpcserver.WriteChanCapacity(n.config.RPC.WebSocketWriteBufferSize),
		)
		wm.SetLogger(wmLogger)
		mux.HandleFunc("/websocket", wm.WebsocketHandler)
		rpcserver.RegisterRPCFuncs(mux, routes, rpcLogger)
		listener, err := rpcserver.Listen(
			listenAddr,
			config.MaxOpenConnections,
		)
		if err != nil {
			return nil, err
		}

		var rootHandler http.Handler = mux
		if n.config.RPC.IsCorsEnabled() {
			corsMiddleware := cors.New(cors.Options{
				AllowedOrigins: n.config.RPC.CORSAllowedOrigins,
				AllowedMethods: n.config.RPC.CORSAllowedMethods,
				AllowedHeaders: n.config.RPC.CORSAllowedHeaders,
			})
			rootHandler = corsMiddleware.Handler(mux)
		}
		if n.config.RPC.IsTLSEnabled() {
			go func() {
				if err := rpcserver.ServeTLS(
					listener,
					rootHandler,
					n.config.RPC.CertFile(),
					n.config.RPC.KeyFile(),
					rpcLogger,
					config,
				); err != nil {
					n.Logger.Error("Error serving server with TLS", "err", err)
				}
			}()
		} else {
			go func() {
				if err := rpcserver.Serve(
					listener,
					rootHandler,
					rpcLogger,
					config,
				); err != nil {
					n.Logger.Error("Error serving server", "err", err)
				}
			}()
		}

		listeners[i] = listener
	}

	// we expose a simplified api over grpc for convenience to app devs
	grpcListenAddr := n.config.RPC.GRPCListenAddress
	if grpcListenAddr != "" {
		config := rpcserver.DefaultConfig()
		config.MaxBodyBytes = n.config.RPC.MaxBodyBytes
		config.MaxHeaderBytes = n.config.RPC.MaxHeaderBytes
		// NOTE: GRPCMaxOpenConnections is used, not MaxOpenConnections
		config.MaxOpenConnections = n.config.RPC.GRPCMaxOpenConnections
		// If necessary adjust global WriteTimeout to ensure it's greater than
		// TimeoutBroadcastTxCommit.
		// See https://github.com/tendermint/tendermint/issues/3435
		if config.WriteTimeout <= n.config.RPC.TimeoutBroadcastTxCommit {
			config.WriteTimeout = n.config.RPC.TimeoutBroadcastTxCommit + 1*time.Second
		}
		listener, err := rpcserver.Listen(grpcListenAddr, config.MaxOpenConnections)
		if err != nil {
			return nil, err
		}
		go func() {
			if err := grpccore.StartGRPCServer(env, listener); err != nil {
				n.Logger.Error("Error starting gRPC server", "err", err)
			}
		}()
		listeners = append(listeners, listener)

	}

	return listeners, nil
}

// startPrometheusServer starts a Prometheus HTTP server, listening for metrics
// collectors on addr.
func (n *Node) startPrometheusServer() *http.Server {
	srv := &http.Server{
		Addr: n.config.Instrumentation.PrometheusListenAddr,
		Handler: promhttp.InstrumentMetricHandler(
			prometheus.DefaultRegisterer, promhttp.HandlerFor(
				prometheus.DefaultGatherer,
				promhttp.HandlerOpts{MaxRequestsInFlight: n.config.Instrumentation.MaxOpenConnections},
			),
		),
		ReadHeaderTimeout: readHeaderTimeout,
	}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// Error starting or closing listener:
			n.Logger.Error("Prometheus HTTP server ListenAndServe", "err", err)
		}
	}()
	return srv
}

// starts a ppro
func (n *Node) startPprofServer() *http.Server {
	srv := &http.Server{
		Addr:              n.config.RPC.PprofListenAddress,
		Handler:           nil,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// Error starting or closing listener:
			n.Logger.Error("pprof HTTP server ListenAndServe", "err", err)
		}
	}()
	return srv
}

// Switch returns the Node's Switch.
func (n *Node) Switch() *p2p.Switch {
	return n.sw
}

// PrivValidator returns the Node's PrivValidator.
// XXX: for convenience only!
func (n *Node) PrivValidator() types.PrivValidator {
	return n.privValidator
}

// ProxyApp returns the Node's AppConns, representing its connections to the AVSI application.
func (n *Node) ProxyApp() proxy.AppConns {
	return n.proxyApp
}

// Config returns the Node's config.
func (n *Node) Config() *cfg.Config {
	return n.config
}

//------------------------------------------------------------------------------

func (n *Node) Listeners() []string {
	return []string{
		fmt.Sprintf("Listener(@%v)", n.config.P2P.ExternalAddress),
	}
}

func (n *Node) IsListening() bool {
	return n.isListening
}

// NodeInfo returns the Node's Info from the Switch.
func (n *Node) NodeInfo() p2p.NodeInfo {
	return n.nodeInfo
}

func makeNodeInfo(
	config *cfg.Config,
	nodeKey *p2p.NodeKey,
	dvsRequestIndexer requestindex.DvsRequestIndexer,
) (p2p.DefaultNodeInfo, error) {
	dvsRequestIndexerStatus := "on"
	if _, ok := dvsRequestIndexer.(*null.DvsRequestIndex); ok {
		dvsRequestIndexerStatus = "off"
	}

	nodeInfo := p2p.DefaultNodeInfo{
		ProtocolVersion: p2p.NewProtocolVersion(
			version.P2PProtocol, // global
			//TODO: change interface
			0,
			0,
		),
		DefaultNodeID: nodeKey.ID(),
		//TODO: get network from config
		Network:  "ID",
		Version:  version.TMCoreSemVer,
		Channels: []byte{},
		Moniker:  config.Moniker,
		Other: p2p.DefaultNodeInfoOther{
			DvsRequestIndex: dvsRequestIndexerStatus,
			RPCAddress:      config.RPC.ListenAddress,
		},
	}

	if config.P2P.PexReactor {
		nodeInfo.Channels = append(nodeInfo.Channels, pex.PexChannel)
	}

	lAddr := config.P2P.ExternalAddress

	if lAddr == "" {
		lAddr = config.P2P.ListenAddress
	}

	nodeInfo.ListenAddr = lAddr

	err := nodeInfo.Validate()
	return nodeInfo, err
}

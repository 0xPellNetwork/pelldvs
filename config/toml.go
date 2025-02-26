package config

import (
	"bytes"
	"path/filepath"
	"strings"
	"text/template"

	cmtos "github.com/0xPellNetwork/pelldvs/libs/os"
)

// DefaultDirPerm is the default permissions used when creating directories.
const DefaultDirPerm = 0700

var configTemplate *template.Template

//var defaultConfigFilePath = filepath.Join(DefaultConfigDir, DefaultConfigFileName)

func init() {
	var err error
	tmpl := template.New("configFileTemplate").Funcs(template.FuncMap{
		"StringsJoin": strings.Join,
	})
	if configTemplate, err = tmpl.Parse(defaultConfigTemplate); err != nil {
		panic(err)
	}
}

/****** these are for production settings ***********/

// EnsureRoot creates the root, config, and data directories if they don't exist,
// and panics if it fails.
func EnsureRoot(rootDir string) {
	if err := cmtos.EnsureDir(rootDir, DefaultDirPerm); err != nil {
		panic(err.Error())
	}
	if err := cmtos.EnsureDir(filepath.Join(rootDir, DefaultConfigDir), DefaultDirPerm); err != nil {
		panic(err.Error())
	}
	if err := cmtos.EnsureDir(filepath.Join(rootDir, DefaultDataDir), DefaultDirPerm); err != nil {
		panic(err.Error())
	}

	configFilePath := filepath.Join(rootDir, defaultConfigFilePath)

	// Write default config file if missing.
	if !cmtos.FileExists(configFilePath) {
		writeDefaultConfigFile(rootDir, configFilePath)
	}
}

// XXX: this func should probably be called by cmd/pelldvs/commands/init.go
// alongside the writing of the genesis.json and priv_validator.json
func writeDefaultConfigFile(root, configFilePath string) {

	config := DefaultConfig()
	config.Instrumentation.Namespace = "pelldvs"

	config.Pell.OperatorECDSAPrivateKeyStorePath = filepath.Join(root, "keys", "operator.ecdsa.key.json")
	config.Pell.OperatorBLSPrivateKeyStorePath = filepath.Join(root, "keys", "operator.bls.key.json")

	WriteConfigFile(configFilePath, config)
}

// WriteConfigFile renders config using the template and writes it to configFilePath.
func WriteConfigFile(configFilePath string, config *Config) {

	//type AllConfig struct {
	//	*Config `mapstructure:",squash"`
	//	Pell    *PellConfig `mapstructure:"pell" `
	//}
	//var pellcfg *PellConfig
	//if pellConfig != nil {
	//	pellcfg = pellConfig
	//} else {
	//	c := DefaultPellConfig()
	//	c.RPCURL = "http://localhost:8545"
	//	c.PellDelegationManagerAddress = "0x4A679253410272dd5232B3Ff7cF5dbB88f295319"
	//	c.PellRegistryRouterAddress = "0xeD343c0f99C89Ed7c3c934A88f90261fD6a9A68b"
	//	c.CentralSchedulerAddress = "0x36C02dA8a0983159322a80FFE9F24b1acfF8B570"
	//	c.PellDVSDirectoryAddress = "0xf5059a5D33d5853360D16C683c16e67980206f36"
	//
	//	pellcfg = c
	//}
	//
	//allConfig := AllConfig{
	//	Config: config,
	//	Pell:   pellcfg,
	//}

	var buffer bytes.Buffer
	if err := configTemplate.Execute(&buffer, config); err != nil {
		panic(err)
	}

	cmtos.MustWriteFile(configFilePath, buffer.Bytes(), 0644)
}

// Note: any changes to the comments/variables/mapstructure
// must be reflected in the appropriate struct in config/config.go
const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

# NOTE: Any path below can be absolute (e.g. "/var/myawesomeapp/data") or
# relative to the home directory (e.g. "data"). The home directory is
# "$HOME/.pelldvs" by default, but could be changed via $PELLDVSHOME env variable
# or --home cmd flag.

# The version of the PellDVS binary that created or
# last modified the config file. Do not modify this.
version = "{{ .BaseConfig.Version }}"

#######################################################################
###                   Main Base Config Options                      ###
#######################################################################

# TCP or UNIX socket address of the ABCI application,
# or the name of an ABCI application compiled in with the PellDVS binary
proxy_app = "{{ .BaseConfig.ProxyApp }}"

# A custom human readable name for this node
moniker = "{{ .BaseConfig.Moniker }}"

# Database backend: goleveldb | cleveldb | boltdb | rocksdb | badgerdb
# * goleveldb (github.com/syndtr/goleveldb - most popular implementation)
#   - pure go
#   - stable
# * cleveldb (uses levigo wrapper)
#   - fast
#   - requires gcc
#   - use cleveldb build tag (go build -tags cleveldb)
# * boltdb (uses etcd's fork of bolt - github.com/etcd-io/bbolt)
#   - EXPERIMENTAL
#   - may be faster is some use-cases (random reads - indexer)
#   - use boltdb build tag (go build -tags boltdb)
# * rocksdb (uses github.com/tecbot/gorocksdb)
#   - EXPERIMENTAL
#   - requires gcc
#   - use rocksdb build tag (go build -tags rocksdb)
# * badgerdb (uses github.com/dgraph-io/badger)
#   - EXPERIMENTAL
#   - use badgerdb build tag (go build -tags badgerdb)
db_backend = "{{ .BaseConfig.DBBackend }}"

# Database directory
db_dir = "{{ js .BaseConfig.DBPath }}"

# Output level for logging, including package level options
log_level = "{{ .BaseConfig.LogLevel }}"

# Output format: 'plain' (colored text) or 'json'
log_format = "{{ .BaseConfig.LogFormat }}"

##### additional base config options #####

# Path to the JSON file containing the initial validator set and other meta data
genesis_file = "{{ js .BaseConfig.Genesis }}"

# Path to the JSON file containing the private key to use as a validator in the consensus protocol
priv_validator_key_file = "{{ js .BaseConfig.PrivValidatorKey }}"

# Path to the JSON file containing the last sign state of a validator
priv_validator_state_file = "{{ js .BaseConfig.PrivValidatorState }}"

# TCP or UNIX socket address for PellDVS to listen on for
# connections from an external PrivValidator process
priv_validator_laddr = "{{ .BaseConfig.PrivValidatorListenAddr }}"

# Path to the JSON file containing the private key to use for node authentication in the p2p protocol
node_key_file = "{{ js .BaseConfig.NodeKey }}"

# Mechanism to connect to the ABCI application: socket | grpc
abci = "{{ .BaseConfig.ABCI }}"

# If true, query the ABCI app on connecting to a new peer
# so the app can decide if we should keep the connection or not
filter_peers = {{ .BaseConfig.FilterPeers }}

#######################################################################
###                 Advanced Configuration Options                  ###
#######################################################################

#######################################################
###       RPC Server Configuration Options          ###
#######################################################
[rpc]

# TCP or UNIX socket address for the RPC server to listen on
laddr = "{{ .RPC.ListenAddress }}"

# A list of origins a cross-domain request can be executed from
# Default value '[]' disables cors support
# Use '["*"]' to allow any origin
cors_allowed_origins = [{{ range .RPC.CORSAllowedOrigins }}{{ printf "%q, " . }}{{end}}]

# A list of methods the client is allowed to use with cross-domain requests
cors_allowed_methods = [{{ range .RPC.CORSAllowedMethods }}{{ printf "%q, " . }}{{end}}]

# A list of non simple headers the client is allowed to use with cross-domain requests
cors_allowed_headers = [{{ range .RPC.CORSAllowedHeaders }}{{ printf "%q, " . }}{{end}}]

# TCP or UNIX socket address for the gRPC server to listen on
# NOTE: This server only supports /broadcast_tx_commit
grpc_laddr = "{{ .RPC.GRPCListenAddress }}"

# Maximum number of simultaneous connections.
# Does not include RPC (HTTP&WebSocket) connections. See max_open_connections
# If you want to accept a larger number than the default, make sure
# you increase your OS limits.
# 0 - unlimited.
# Should be < {ulimit -Sn} - {MaxNumInboundPeers} - {MaxNumOutboundPeers} - {N of wal, db and other open files}
# 1024 - 40 - 10 - 50 = 924 = ~900
grpc_max_open_connections = {{ .RPC.GRPCMaxOpenConnections }}

# Activate unsafe RPC commands like /dial_seeds and /unsafe_flush_mempool
unsafe = {{ .RPC.Unsafe }}

# Maximum number of simultaneous connections (including WebSocket).
# Does not include gRPC connections. See grpc_max_open_connections
# If you want to accept a larger number than the default, make sure
# you increase your OS limits.
# 0 - unlimited.
# Should be < {ulimit -Sn} - {MaxNumInboundPeers} - {MaxNumOutboundPeers} - {N of wal, db and other open files}
# 1024 - 40 - 10 - 50 = 924 = ~900
max_open_connections = {{ .RPC.MaxOpenConnections }}

# Maximum number of unique clientIDs that can /subscribe
# If you're using /broadcast_tx_commit, set to the estimated maximum number
# of broadcast_tx_commit calls per block.
max_subscription_clients = {{ .RPC.MaxSubscriptionClients }}

# Maximum number of unique queries a given client can /subscribe to
# If you're using GRPC (or Local RPC client) and /broadcast_tx_commit, set to
# the estimated # maximum number of broadcast_tx_commit calls per block.
max_subscriptions_per_client = {{ .RPC.MaxSubscriptionsPerClient }}

# Experimental parameter to specify the maximum number of events a node will
# buffer, per subscription, before returning an error and closing the
# subscription. Must be set to at least 100, but higher values will accommodate
# higher event throughput rates (and will use more memory).
experimental_subscription_buffer_size = {{ .RPC.SubscriptionBufferSize }}

# Experimental parameter to specify the maximum number of RPC responses that
# can be buffered per WebSocket client. If clients cannot read from the
# WebSocket endpoint fast enough, they will be disconnected, so increasing this
# parameter may reduce the chances of them being disconnected (but will cause
# the node to use more memory).
#
# Must be at least the same as "experimental_subscription_buffer_size",
# otherwise connections could be dropped unnecessarily. This value should
# ideally be somewhat higher than "experimental_subscription_buffer_size" to
# accommodate non-subscription-related RPC responses.
experimental_websocket_write_buffer_size = {{ .RPC.WebSocketWriteBufferSize }}

# If a WebSocket client cannot read fast enough, at present we may
# silently drop events instead of generating an error or disconnecting the
# client.
#
# Enabling this experimental parameter will cause the WebSocket connection to
# be closed instead if it cannot read fast enough, allowing for greater
# predictability in subscription behavior.
experimental_close_on_slow_client = {{ .RPC.CloseOnSlowClient }}

# How long to wait for a tx to be committed during /broadcast_tx_commit.
# WARNING: Using a value larger than 10s will result in increasing the
# global HTTP write timeout, which applies to all connections and endpoints.
# See https://github.com/tendermint/tendermint/issues/3435
timeout_broadcast_tx_commit = "{{ .RPC.TimeoutBroadcastTxCommit }}"

# Maximum number of requests that can be sent in a batch
# If the value is set to '0' (zero-value), then no maximum batch size will be
# enforced for a JSON-RPC batch request.
max_request_batch_size = {{ .RPC.MaxRequestBatchSize }}

# Maximum size of request body, in bytes
max_body_bytes = {{ .RPC.MaxBodyBytes }}

# Maximum size of request header, in bytes
max_header_bytes = {{ .RPC.MaxHeaderBytes }}

# The path to a file containing certificate that is used to create the HTTPS server.
# Might be either absolute path or path related to PellDVS's config directory.
# If the certificate is signed by a certificate authority,
# the certFile should be the concatenation of the server's certificate, any intermediates,
# and the CA's certificate.
# NOTE: both tls_cert_file and tls_key_file must be present for PellDVS to create HTTPS server.
# Otherwise, HTTP server is run.
tls_cert_file = "{{ .RPC.TLSCertFile }}"

# The path to a file containing matching private key that is used to create the HTTPS server.
# Might be either absolute path or path related to PellDVS's config directory.
# NOTE: both tls-cert-file and tls-key-file must be present for PellDVS to create HTTPS server.
# Otherwise, HTTP server is run.
tls_key_file = "{{ .RPC.TLSKeyFile }}"

# pprof listen address (https://golang.org/pkg/net/http/pprof)
pprof_laddr = "{{ .RPC.PprofListenAddress }}"

#######################################################
###           P2P Configuration Options             ###
#######################################################
[p2p]

# Address to listen for incoming connections
laddr = "{{ .P2P.ListenAddress }}"

# Address to advertise to peers for them to dial. If empty, will use the same
# port as the laddr, and will introspect on the listener to figure out the
# address. IP and port are required. Example: 159.89.10.97:26656
external_address = "{{ .P2P.ExternalAddress }}"

# Comma separated list of seed nodes to connect to
seeds = "{{ .P2P.Seeds }}"

# Comma separated list of nodes to keep persistent connections to
persistent_peers = "{{ .P2P.PersistentPeers }}"

# Path to address book
addr_book_file = "{{ js .P2P.AddrBook }}"

# Set true for strict address routability rules
# Set false for private or local networks
addr_book_strict = {{ .P2P.AddrBookStrict }}

# Maximum number of inbound peers
max_num_inbound_peers = {{ .P2P.MaxNumInboundPeers }}

# Maximum number of outbound peers to connect to, excluding persistent peers
max_num_outbound_peers = {{ .P2P.MaxNumOutboundPeers }}

# List of node IDs, to which a connection will be (re)established ignoring any existing limits
unconditional_peer_ids = "{{ .P2P.UnconditionalPeerIDs }}"

# Maximum pause when redialing a persistent peer (if zero, exponential backoff is used)
persistent_peers_max_dial_period = "{{ .P2P.PersistentPeersMaxDialPeriod }}"

# Time to wait before flushing messages out on the connection
flush_throttle_timeout = "{{ .P2P.FlushThrottleTimeout }}"

# Maximum size of a message packet payload, in bytes
max_packet_msg_payload_size = {{ .P2P.MaxPacketMsgPayloadSize }}

# Rate at which packets can be sent, in bytes/second
send_rate = {{ .P2P.SendRate }}

# Rate at which packets can be received, in bytes/second
recv_rate = {{ .P2P.RecvRate }}

# Set true to enable the peer-exchange reactor
pex = {{ .P2P.PexReactor }}

# Seed mode, in which node constantly crawls the network and looks for
# peers. If another node asks it for addresses, it responds and disconnects.
#
# Does not work if the peer-exchange reactor is disabled.
seed_mode = {{ .P2P.SeedMode }}

# Comma separated list of peer IDs to keep private (will not be gossiped to other peers)
private_peer_ids = "{{ .P2P.PrivatePeerIDs }}"

# Toggle to disable guard against peers connecting from the same ip.
allow_duplicate_ip = {{ .P2P.AllowDuplicateIP }}

# Peer connection configuration.
handshake_timeout = "{{ .P2P.HandshakeTimeout }}"
dial_timeout = "{{ .P2P.DialTimeout }}"

#######################################################
###   Transaction Indexer Configuration Options     ###
#######################################################
[dvs_request_index]

# What indexer to use for transactions
#
# The application will set which txs to index. In some cases a node operator will be able
# to decide which txs to index based on configuration set in the application.
#
# Options:
#   1) "null"
#   2) "kv" (default) - the simplest possible indexer, backed by key-value storage (defaults to levelDB; see DBBackend).
# 		- When "kv" is chosen "tx.height" and "tx.hash" will always be indexed.
#   3) "psql" - the indexer services backed by PostgreSQL.
# When "kv" or "psql" is chosen "tx.height" and "tx.hash" will always be indexed.
indexer = "{{ .DVSRequestIndex.Indexer }}"

# The PostgreSQL connection configuration, the connection format:
#   postgresql://<user>:<password>@<host>:<port>/<db>?<opts>
psql-conn = "{{ .DVSRequestIndex.PsqlConn }}"

#######################################################
###       Instrumentation Configuration Options     ###
#######################################################
[instrumentation]

# When true, Prometheus metrics are served under /metrics on
# PrometheusListenAddr.
# Check out the documentation for the list of available metrics.
prometheus = {{ .Instrumentation.Prometheus }}

# Address to listen for Prometheus collector(s) connections
prometheus_listen_addr = "{{ .Instrumentation.PrometheusListenAddr }}"

# Maximum number of simultaneous connections.
# If you want to accept a larger number than the default, make sure
# you increase your OS limits.
# 0 - unlimited.
max_open_connections = {{ .Instrumentation.MaxOpenConnections }}

# Instrumentation namespace
namespace = "{{ .Instrumentation.Namespace }}"

#######################################################
###       Pell Configuration Options                ###
#######################################################

# Pell Configuration
[pell]

# Aggregator RPC URL
aggregator_rpc_url = "{{ .Pell.AggregatorRPCURL }}"

# path to the file containing the private key for the operator ECDSA key
operator_ecdsa_private_key_store_path = "{{ .Pell.OperatorECDSAPrivateKeyStorePath }}"

# Path to the file containing the private key for the operator BLS key
operator_bls_private_key_store_path = "{{ .Pell.OperatorBLSPrivateKeyStorePath }}"

# Chain config path
interfactor_config_path = "{{ .Pell.InteractorConfigPath }}"
`

package coretypes

import (
	"encoding/json"

	avsi "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/libs/bytes"
	"github.com/0xPellNetwork/pelldvs/p2p"
)

// request result
type ResultRequest struct {
	Code      uint32         `json:"code"`
	Data      bytes.HexBytes `json:"data"`
	Log       string         `json:"log"`
	Codespace string         `json:"codespace"`
}

type ResultDvsRequest struct {
	DvsRequest                 *avsi.DVSRequest                 `json:"dvs_request"`
	DvsResponse                *avsi.DVSResponse                `json:"dvs_response"`
	ResponseProcessDvsRequest  *avsi.ResponseProcessDVSRequest  `json:"response_dvs_request"`
	ResponseProcessDVSResponse *avsi.ResponseProcessDVSResponse `json:"response_dvs_response"`
	Hash                       bytes.HexBytes                   `json:"hash,omitempty"`
}

// Result of searching for dvs request
type ResultDvsRequestSearch struct {
	DvsRequests []*ResultDvsRequest `json:"dvs_requests"`
	TotalCount  int                 `json:"total_count"`
}

type ResultRequestDvsAsync struct {
	Hash bytes.HexBytes `json:"hash"`
}

// Info avsi msg
type ResultAVSIInfo struct {
	Response avsi.ResponseInfo `json:"response"`
}

// avsi
type ResultAVSIQuery struct {
	Response avsi.ResponseQuery `json:"response"`
}

// ResultGenesisChunk is the output format for the chunked/paginated
// interface. These chunks are produced by converting the genesis
// document to JSON and then splitting the resulting payload into
// 16 megabyte blocks and then base64 encoding each block.
type ResultGenesisChunk struct {
	ChunkNumber int    `json:"chunk"`
	TotalChunks int    `json:"total"`
	Data        string `json:"data"`
}

// Info about peer connections
type ResultNetInfo struct {
	Listening bool     `json:"listening"`
	Listeners []string `json:"listeners"`
	NPeers    int      `json:"n_peers"`
	Peers     []Peer   `json:"peers"`
}

// Log from dialing seeds
type ResultDialSeeds struct {
	Log string `json:"log"`
}

// Log from dialing peers
type ResultDialPeers struct {
	Log string `json:"log"`
}

// A peer
type Peer struct {
	NodeInfo         p2p.DefaultNodeInfo  `json:"node_info"`
	IsOutbound       bool                 `json:"is_outbound"`
	ConnectionStatus p2p.ConnectionStatus `json:"connection_status"`
	RemoteIP         string               `json:"remote_ip"`
}

// UNSTABLE
type PeerStateInfo struct {
	NodeAddress string          `json:"node_address"`
	PeerState   json.RawMessage `json:"peer_state"`
}

// empty results
type (
	// ResultUnsafeFlushMempool struct{}
	// ResultUnsafeProfile      struct{}
	// ResultSubscribe          struct{}
	// ResultUnsubscribe        struct{}
	ResultHealth struct{}
)

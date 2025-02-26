package core

import (
	"fmt"
	"time"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	cfg "github.com/0xPellNetwork/pelldvs/config"
	"github.com/0xPellNetwork/pelldvs/p2p"
	"github.com/0xPellNetwork/pelldvs/proxy"
	"github.com/0xPellNetwork/pelldvs/security"
	"github.com/0xPellNetwork/pelldvs/state/requestindex"
)

const (
	// see README
	defaultPerPage = 30
	maxPerPage     = 100

	// SubscribeTimeout is the maximum time we wait to subscribe for an event.
	// must be less than the server's write timeout (see rpcserver.DefaultConfig)
	SubscribeTimeout = 5 * time.Second
)

//----------------------------------------------
// These interfaces are used by RPC and must be thread safe

type transport interface {
	Listeners() []string
	IsListening() bool
	NodeInfo() p2p.NodeInfo
}

type peers interface {
	AddPersistentPeers([]string) error
	AddUnconditionalPeerIDs([]string) error
	AddPrivatePeerIDs([]string) error
	DialPeersAsync([]string) error
	Peers() p2p.IPeerSet
}

// ----------------------------------------------
// Environment contains objects and interfaces used by the RPC. It is expected
// to be setup once during startup.
type Environment struct {
	// external, thread safe interfaces
	ProxyAppQuery proxy.AppConnQuery
	DVSReactor    security.DVSReactor
	// ProxyAppMempool proxy.AppConnMempool

	P2PPeers     peers
	P2PTransport transport

	DvsRequestIndexer requestindex.DvsRequestIndexer

	// objects
	//EventBus *types.EventBus // thread safe

	Logger log.Logger
	Config cfg.RPCConfig
}

//----------------------------------------------

func validatePage(pagePtr *int, perPage, totalCount int) (int, error) {
	if perPage < 1 {
		panic(fmt.Sprintf("zero or negative perPage: %d", perPage))
	}

	if pagePtr == nil { // no page parameter
		return 1, nil
	}

	pages := ((totalCount - 1) / perPage) + 1
	if pages == 0 {
		pages = 1 // one page (even if it's empty)
	}
	page := *pagePtr
	if page <= 0 || page > pages {
		return 1, fmt.Errorf("page should be within [1, %d] range, given %d", pages, page)
	}

	return page, nil
}

func (env *Environment) validatePerPage(perPagePtr *int) int {
	if perPagePtr == nil { // no per_page parameter
		return defaultPerPage
	}

	perPage := *perPagePtr
	if perPage < 1 {
		return defaultPerPage
	} else if perPage > maxPerPage {
		return maxPerPage
	}
	return perPage
}

func validateSkipCount(page, perPage int) int {
	skipCount := (page - 1) * perPage
	if skipCount < 0 {
		return 0
	}

	return skipCount
}

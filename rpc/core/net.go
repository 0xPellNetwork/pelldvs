package core

import (
	"errors"
	"fmt"
	"strings"

	"github.com/0xPellNetwork/pelldvs/p2p"
	ctypes "github.com/0xPellNetwork/pelldvs/rpc/core/types"
	rpctypes "github.com/0xPellNetwork/pelldvs/rpc/jsonrpc/types"
)

// NetInfo returns network info.
func (env *Environment) NetInfo(*rpctypes.Context) (*ctypes.ResultNetInfo, error) {
	peersList := env.P2PPeers.Peers().List()
	peers := make([]ctypes.Peer, 0, len(peersList))
	for _, peer := range peersList {
		nodeInfo, ok := peer.NodeInfo().(p2p.DefaultNodeInfo)
		if !ok {
			return nil, fmt.Errorf("peer.NodeInfo() is not DefaultNodeInfo")
		}
		peers = append(peers, ctypes.Peer{
			NodeInfo:         nodeInfo,
			IsOutbound:       peer.IsOutbound(),
			ConnectionStatus: peer.Status(),
			RemoteIP:         peer.RemoteIP().String(),
		})
	}
	// TODO: Should we include PersistentPeers and Seeds in here?
	// PRO: useful info
	// CON: privacy
	return &ctypes.ResultNetInfo{
		Listening: env.P2PTransport.IsListening(),
		Listeners: env.P2PTransport.Listeners(),
		NPeers:    len(peers),
		Peers:     peers,
	}, nil
}

// UnsafeDialSeeds dials the given seeds (comma-separated id@IP:PORT).
func (env *Environment) UnsafeDialSeeds(_ *rpctypes.Context, seeds []string) (*ctypes.ResultDialSeeds, error) {
	if len(seeds) == 0 {
		return &ctypes.ResultDialSeeds{}, errors.New("no seeds provided")
	}
	env.Logger.Info("DialSeeds", "seeds", seeds)
	if err := env.P2PPeers.DialPeersAsync(seeds); err != nil {
		return &ctypes.ResultDialSeeds{}, err
	}
	return &ctypes.ResultDialSeeds{Log: "Dialing seeds in progress. See /net_info for details"}, nil
}

// UnsafeDialPeers dials the given peers (comma-separated id@IP:PORT),
// optionally making them persistent.
func (env *Environment) UnsafeDialPeers(
	_ *rpctypes.Context,
	peers []string,
	persistent, unconditional, private bool,
) (*ctypes.ResultDialPeers, error) {
	if len(peers) == 0 {
		return &ctypes.ResultDialPeers{}, errors.New("no peers provided")
	}

	ids, err := getIDs(peers)
	if err != nil {
		return &ctypes.ResultDialPeers{}, err
	}

	env.Logger.Info("DialPeers", "peers", peers, "persistent",
		persistent, "unconditional", unconditional, "private", private)

	if persistent {
		if err := env.P2PPeers.AddPersistentPeers(peers); err != nil {
			return &ctypes.ResultDialPeers{}, err
		}
	}

	if private {
		if err := env.P2PPeers.AddPrivatePeerIDs(ids); err != nil {
			return &ctypes.ResultDialPeers{}, err
		}
	}

	if unconditional {
		if err := env.P2PPeers.AddUnconditionalPeerIDs(ids); err != nil {
			return &ctypes.ResultDialPeers{}, err
		}
	}

	if err := env.P2PPeers.DialPeersAsync(peers); err != nil {
		return &ctypes.ResultDialPeers{}, err
	}

	return &ctypes.ResultDialPeers{Log: "Dialing peers in progress. See /net_info for details"}, nil
}

func getIDs(peers []string) ([]string, error) {
	ids := make([]string, 0, len(peers))

	for _, peer := range peers {

		spl := strings.Split(peer, "@")
		if len(spl) != 2 {
			return nil, p2p.ErrNetAddressNoID{Addr: peer}
		}
		ids = append(ids, spl[0])

	}
	return ids, nil
}

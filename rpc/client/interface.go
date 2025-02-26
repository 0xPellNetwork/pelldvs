package client

/*
The client package provides a general purpose interface (Client) for connecting
to a PellDVS node, as well as higher-level functionality.

The main implementation for production code is client.HTTP, which
connects via http to the jsonrpc interface of the PellDVS node.

For connecting to a node running in the same process (eg. when
compiling the abci app in the same process), you can use the client.Local
implementation.

For mocking out server responses during testing to see behavior for
arbitrary return values, use the mock package.

In addition to the Client interface, which should be used externally
for maximum flexibility and testability, and two implementations,
this package also provides helper functions that work on any Client
implementation.
*/

import (
	"context"

	"github.com/0xPellNetwork/pelldvs/libs/service"
	ctypes "github.com/0xPellNetwork/pelldvs/rpc/core/types"
)

// Client wraps most important rpc calls a client would make if you want to
// listen for events, test if it also implements events.EventSwitch.
type Client interface {
	service.Service
}

// HistoryClient provides access to data from genesis to now in large chunks.
type HistoryClient interface {
	// Genesis(context.Context) (*ctypes.ResultGenesis, error)
	GenesisChunked(context.Context, uint) (*ctypes.ResultGenesisChunk, error)
	//BlockchainInfo(ctx context.Context, minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error)
}

// NetworkClient is general info about the network state. May not be needed
// usually.
type NetworkClient interface {
	NetInfo(context.Context) (*ctypes.ResultNetInfo, error)
	//ConsensusState(context.Context) (*ctypes.ResultConsensusState, error)
	//ConsensusParams(ctx context.Context, height *int64) (*ctypes.ResultConsensusParams, error)
	Health(context.Context) (*ctypes.ResultHealth, error)
}

// EventsClient is reactive, you can subscribe to any message, given the proper
// string. see pelldvs/types/events.go
type EventsClient interface {
	// Subscribe subscribes given subscriber to query. Returns a channel with
	// cap=1 onto which events are published. An error is returned if it fails to
	// subscribe. outCapacity can be used optionally to set capacity for the
	// channel. Channel is never closed to prevent accidental reads.
	//
	// ctx cannot be used to unsubscribe. To unsubscribe, use either Unsubscribe
	// or UnsubscribeAll.
	//Subscribe(ctx context.Context, subscriber, query string, outCapacity ...int) (out <-chan ctypes.ResultEvent, err error)
	// Unsubscribe unsubscribes given subscriber from query.
	Unsubscribe(ctx context.Context, subscriber, query string) error
	// UnsubscribeAll unsubscribes given subscriber from all the queries.
	UnsubscribeAll(ctx context.Context, subscriber string) error
}

// DVSClient
type DVSClient interface {
	RequestDVS(
		ctx context.Context,
		data []byte,
		height int64,
		chainid int64,
		groupNumbers []uint32,
		groupThresholdPercentages []uint32,
	) (*ctypes.ResultRequest, error)

	RequestDVSAsync(
		ctx context.Context,
		data []byte,
		height int64,
		chainid int64,
		groupNumbers []uint32,
		groupThresholdPercentages []uint32,
	) (*ctypes.ResultRequestDvsAsync, error)

	QueryRequest(ctx context.Context, hash string) (*ctypes.ResultDvsRequest, error)
}

// RemoteClient is a Client, which can also return the remote network address.
type RemoteClient interface {
	Client

	// Remote returns the remote network address in a string form.
	Remote() string
}

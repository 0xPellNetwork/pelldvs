package local

import (
	"context"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	nm "github.com/0xPellNetwork/pelldvs/node"
	"github.com/0xPellNetwork/pelldvs/rpc/core"
	ctypes "github.com/0xPellNetwork/pelldvs/rpc/core/types"
	rpctypes "github.com/0xPellNetwork/pelldvs/rpc/jsonrpc/types"
)

/*
Local is a Client implementation that directly executes the rpc
functions on a given node, without going through HTTP or GRPC.

This implementation is useful for:

* Running tests against a node in-process without the overhead
of going through an http server
* Communication between an AVSI app and PellDVS when they
are compiled in process.

For real clients, you probably want to use client.HTTP.  For more
powerful control during testing, you probably want the "client/mock" package.

You can subscribe for any event published by PellDVS using Subscribe method.
Note delivery is best-effort. If you don't read events fast enough, PellDVS
might cancel the subscription. The client will attempt to resubscribe (you
don't need to do anything). It will keep trying indefinitely with exponential
backoff (10ms -> 20ms -> 40ms) until successful.
*/
type Local struct {
	// *types.EventBus
	Logger log.Logger
	ctx    *rpctypes.Context
	env    *core.Environment
}

// NewLocal configures a client that calls the Node directly.
func New(node *nm.Node) *Local {
	env, err := node.ConfigureRPC()
	if err != nil {
		node.Logger.Error("Error configuring RPC", "err", err)
	}
	return &Local{
		// EventBus: node.EventBus(),
		Logger: log.NewNopLogger(),
		ctx:    &rpctypes.Context{},
		env:    env,
	}
}

// var _ rpcclient.Client = (*Local)(nil)

func (c *Local) IsRunning() bool {
	return false
}

// SetLogger allows to set a logger on the client.
func (c *Local) SetLogger(l log.Logger) {
	c.Logger = l
}

func (c *Local) NetInfo(context.Context) (*ctypes.ResultNetInfo, error) {
	return c.env.NetInfo(c.ctx)
}

func (c *Local) Health(context.Context) (*ctypes.ResultHealth, error) {
	return c.env.Health(c.ctx)
}

func (c *Local) DialSeeds(_ context.Context, seeds []string) (*ctypes.ResultDialSeeds, error) {
	return c.env.UnsafeDialSeeds(c.ctx, seeds)
}

func (c *Local) DialPeers(
	_ context.Context,
	peers []string,
	persistent,
	unconditional,
	private bool,
) (*ctypes.ResultDialPeers, error) {
	return c.env.UnsafeDialPeers(c.ctx, peers, persistent, unconditional, private)
}

func (c *Local) GenesisChunked(_ context.Context, id uint) (*ctypes.ResultGenesisChunk, error) {
	return &ctypes.ResultGenesisChunk{}, nil
}

func (c *Local) RequestDVS(
	_ context.Context,
	data []byte,
	height int64,
	chainid int64,
	groupNumbers []uint32,
	groupThresholdPercentages []uint32,
) (*ctypes.ResultRequest, error) {
	return c.env.RequestDVS(c.ctx, data, height, chainid, groupNumbers, groupThresholdPercentages)
}

func (c *Local) RequestDVSAsync(
	_ context.Context,
	data []byte,
	height int64,
	chainid int64,
	groupNumbers []uint32,
	groupThresholdPercentages []uint32,
) (*ctypes.ResultRequestDvsAsync, error) {
	return c.env.RequestDVSAsync(c.ctx, data, height, chainid, groupNumbers, groupThresholdPercentages)
}

func (c *Local) QueryRequest(hash string) (*ctypes.ResultDvsRequest, error) {
	return c.env.QueryRequest(c.ctx, hash)
}

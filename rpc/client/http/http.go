package http

import (
	"context"
	"net/http"
	"time"

	rpcclient "github.com/0xPellNetwork/pelldvs/rpc/client"
	ctypes "github.com/0xPellNetwork/pelldvs/rpc/core/types"
	jsonrpcclient "github.com/0xPellNetwork/pelldvs/rpc/jsonrpc/client"
)

/*
HTTP is a Client implementation that communicates with a PellDVS node over
JSON RPC and WebSockets.

This is the main implementation you probably want to use in production code.
There are other implementations when calling the PellDVS node in-process
(Local), or when you want to mock out the server for test code (mock).

You can subscribe for any event published by PellDVS using Subscribe method.
Note delivery is best-effort. If you don't read events fast enough or network is
slow, PellDVS might cancel the subscription. The client will attempt to
resubscribe (you don't need to do anything). It will keep trying every second
indefinitely until successful.

Request batching is available for JSON RPC requests over HTTP, which conforms to
the JSON RPC specification (https://www.jsonrpc.org/specification#batch). See
the example for more details.

Example:

	c, err := New("http://192.168.1.10:26657", "/websocket")
	if err != nil {
		// handle error
	}

	// call Start/Stop if you're subscribing to events
	err = c.Start()
	if err != nil {
		// handle error
	}
	defer c.Stop()

	res, err := c.Status()
	if err != nil {
		// handle error
	}

	// handle result
*/
type HTTP struct {
	remote string
	rpc    *jsonrpcclient.Client

	*baseRPCClient
	//*WSEvents
}

// BatchHTTP provides the same interface as `HTTP`, but allows for batching of
// requests (as per https://www.jsonrpc.org/specification#batch). Do not
// instantiate directly - rather use the HTTP.NewBatch() method to create an
// instance of this struct.
//
// Batching of HTTP requests is thread-safe in the sense that multiple
// goroutines can each create their own batches and send them using the same
// HTTP client. Multiple goroutines could also enqueue transactions in a single
// batch, but ordering of transactions in the batch cannot be guaranteed in such
// an example.
type BatchHTTP struct {
	rpcBatch *jsonrpcclient.RequestBatch
	*baseRPCClient
}

// rpcClient is an internal interface to which our RPC clients (batch and
// non-batch) must conform. Acts as an additional code-level sanity check to
// make sure the implementations stay coherent.
type rpcClient interface {
	rpcclient.HistoryClient
	rpcclient.NetworkClient
}

// baseRPCClient implements the basic RPC method logic without the actual
// underlying RPC call functionality, which is provided by `caller`.
type baseRPCClient struct {
	caller jsonrpcclient.Caller
}

var (
	_ rpcClient = (*HTTP)(nil)
	_ rpcClient = (*BatchHTTP)(nil)
	_ rpcClient = (*baseRPCClient)(nil)
)

//-----------------------------------------------------------------------------
// HTTP

// New takes a remote endpoint in the form <protocol>://<host>:<port> and
// the websocket path (which always seems to be "/websocket")
// An error is returned on invalid remote. The function panics when remote is nil.
func New(remote, wsEndpoint string) (*HTTP, error) {
	httpClient, err := jsonrpcclient.DefaultHTTPClient(remote)
	if err != nil {
		return nil, err
	}
	return NewWithClient(remote, wsEndpoint, httpClient)
}

// Create timeout enabled http client
func NewWithTimeout(remote, wsEndpoint string, timeout uint) (*HTTP, error) {
	httpClient, err := jsonrpcclient.DefaultHTTPClient(remote)
	if err != nil {
		return nil, err
	}
	httpClient.Timeout = time.Duration(timeout) * time.Second
	return NewWithClient(remote, wsEndpoint, httpClient)
}

// NewWithClient allows for setting a custom http client (See New).
// An error is returned on invalid remote. The function panics when remote is nil.
func NewWithClient(remote, wsEndpoint string, client *http.Client) (*HTTP, error) {
	if client == nil {
		panic("nil http.Client provided")
	}

	rc, err := jsonrpcclient.NewWithHTTPClient(remote, client)
	if err != nil {
		return nil, err
	}

	httpClient := &HTTP{
		rpc:           rc,
		remote:        remote,
		baseRPCClient: &baseRPCClient{caller: rc},
	}

	return httpClient, nil
}

// Remote returns the remote network address in a string form.
func (c *HTTP) Remote() string {
	return c.remote
}

// NewBatch creates a new batch client for this HTTP client.
func (c *HTTP) NewBatch() *BatchHTTP {
	rpcBatch := c.rpc.NewRequestBatch()
	return &BatchHTTP{
		rpcBatch: rpcBatch,
		baseRPCClient: &baseRPCClient{
			caller: rpcBatch,
		},
	}
}

//-----------------------------------------------------------------------------
// BatchHTTP

// Send is a convenience function for an HTTP batch that will trigger the
// compilation of the batched requests and send them off using the client as a
// single request. On success, this returns a list of the deserialized results
// from each request in the sent batch.
func (b *BatchHTTP) Send(ctx context.Context) ([]interface{}, error) {
	return b.rpcBatch.Send(ctx)
}

// Clear will empty out this batch of requests and return the number of requests
// that were cleared out.
func (b *BatchHTTP) Clear() int {
	return b.rpcBatch.Clear()
}

// Count returns the number of enqueued requests waiting to be sent.
func (b *BatchHTTP) Count() int {
	return b.rpcBatch.Count()
}

func (c *baseRPCClient) NetInfo(ctx context.Context) (*ctypes.ResultNetInfo, error) {
	result := new(ctypes.ResultNetInfo)
	_, err := c.caller.Call(ctx, "net_info", map[string]interface{}{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *baseRPCClient) Health(ctx context.Context) (*ctypes.ResultHealth, error) {
	result := new(ctypes.ResultHealth)
	_, err := c.caller.Call(ctx, "health", map[string]interface{}{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *baseRPCClient) GenesisChunked(ctx context.Context, id uint) (*ctypes.ResultGenesisChunk, error) {
	result := new(ctypes.ResultGenesisChunk)
	_, err := c.caller.Call(ctx, "genesis_chunked", map[string]interface{}{"chunk": id}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *baseRPCClient) RequestDVS(
	ctx context.Context,
	data []byte,
	height int64,
	chainid int64,
	groupNumbers []uint32,
	groupThresholdPercentages []uint32,
) (*ctypes.ResultRequest, error) {
	result := new(ctypes.ResultRequest)
	_, err := c.caller.Call(ctx, "request_dvs", map[string]interface{}{
		"data":                        data,
		"height":                      height,
		"chainid":                     chainid,
		"group_numbers":               groupNumbers,
		"group_threshold_percentages": groupThresholdPercentages,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *baseRPCClient) RequestDVSAsync(
	ctx context.Context,
	data []byte,
	height int64,
	chainid int64,
	groupNumbers []uint32,
	groupThresholdPercentages []uint32,
) (*ctypes.ResultRequestDvsAsync, error) {
	result := new(ctypes.ResultRequestDvsAsync)
	_, err := c.caller.Call(ctx, "request_dvs_async", map[string]interface{}{
		"data":                        data,
		"height":                      height,
		"chainid":                     chainid,
		"group_numbers":               groupNumbers,
		"group_threshold_percentages": groupThresholdPercentages,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *baseRPCClient) QueryRequest(ctx context.Context, hash string) (*ctypes.ResultDvsRequest, error) {
	result := new(ctypes.ResultDvsRequest)
	_, err := c.caller.Call(ctx, "query_request", map[string]interface{}{
		"hash": hash,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *baseRPCClient) SearchRequest(ctx context.Context, query string, pagePtr, perPagePtr *int) (*ctypes.ResultDvsRequestSearch, error) {
	result := new(ctypes.ResultDvsRequestSearch)
	_, err := c.caller.Call(ctx, "search_request", map[string]interface{}{
		"query":    query,
		"page":     pagePtr,
		"per_page": perPagePtr,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

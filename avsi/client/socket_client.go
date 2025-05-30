package avsicli

import (
	"bufio"
	"container/list"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/0xPellNetwork/pelldvs/avsi/types"
	cmtnet "github.com/0xPellNetwork/pelldvs/libs/net"
	"github.com/0xPellNetwork/pelldvs/libs/service"
	"github.com/0xPellNetwork/pelldvs/libs/timer"
)

const (
	reqQueueSize    = 256 // TODO make configurable
	flushThrottleMS = 20  // Don't wait longer than...
)

// socketClient is the client side implementation of the Tendermint
// Socket Protocol (TSP). It is used by an instance of Tendermint to pass
// AVSI requests to an out of process application running the socketServer.
//
// This is goroutine-safe. All calls are serialized to the server through an unbuffered queue. The socketClient
// tracks responses and expects them to respect the order of the requests sent.
type socketClient struct {
	service.BaseService

	addr        string
	mustConnect bool
	conn        net.Conn

	reqQueue   chan *ReqRes
	flushTimer *timer.ThrottleTimer

	mtx     sync.Mutex
	err     error
	reqSent *list.List                            // list of requests sent, waiting for response
	resCb   func(*types.Request, *types.Response) // called on all requests, if set.
}

var _ Client = (*socketClient)(nil)

// NewSocketClient creates a new socket client, which connects to a given
// address. If mustConnect is true, the client will return an error upon start
// if it fails to connect else it will continue to retry.
func NewSocketClient(addr string, mustConnect bool) Client {
	cli := &socketClient{
		reqQueue:    make(chan *ReqRes, reqQueueSize),
		flushTimer:  timer.NewThrottleTimer("socketClient", flushThrottleMS),
		mustConnect: mustConnect,

		addr:    addr,
		reqSent: list.New(),
		resCb:   nil,
	}
	cli.BaseService = *service.NewBaseService(nil, "socketClient", cli)
	return cli
}

// OnStart implements Service by connecting to the server and spawning reading
// and writing goroutines.
func (cli *socketClient) OnStart() error {
	var (
		err  error
		conn net.Conn
	)

	for {
		conn, err = cmtnet.Connect(cli.addr)
		if err != nil {
			if cli.mustConnect {
				return err
			}
			cli.Logger.Error(fmt.Sprintf("avsi.socketClient failed to connect to %v.  Retrying after %vs...",
				cli.addr, dialRetryIntervalSeconds), "err", err)
			time.Sleep(time.Second * dialRetryIntervalSeconds)
			continue
		}
		cli.conn = conn

		go cli.sendRequestsRoutine(conn)
		go cli.recvResponseRoutine(conn)

		return nil
	}
}

// OnStop implements Service by closing connection and flushing all queues.
func (cli *socketClient) OnStop() {
	if cli.conn != nil {
		cli.conn.Close()
	}

	cli.flushQueue()
	cli.flushTimer.Stop()
}

// Error returns an error if the client was stopped abruptly.
func (cli *socketClient) Error() error {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()
	return cli.err
}

//----------------------------------------

func (cli *socketClient) sendRequestsRoutine(conn io.Writer) {
	w := bufio.NewWriter(conn)
	for {
		select {
		case reqres := <-cli.reqQueue:
			// N.B. We must enqueue before sending out the request, otherwise the
			// server may reply before we do it, and the receiver will fail for an
			// unsolicited reply.
			cli.trackRequest(reqres)

			err := types.WriteMessage(reqres.Request, w)
			if err != nil {
				cli.stopForError(fmt.Errorf("write to buffer: %w", err))
				return
			}

			// If it's a flush request, flush the current buffer.
			if _, ok := reqres.Request.Value.(*types.Request_Flush); ok {
				err = w.Flush()
				if err != nil {
					cli.stopForError(fmt.Errorf("flush buffer: %w", err))
					return
				}
			}
		case <-cli.flushTimer.Ch: // flush queue
			select {
			case cli.reqQueue <- NewReqRes(types.ToRequestFlush()):
			default:
				// Probably will fill the buffer, or retry later.
			}
		case <-cli.Quit():
			return
		}
	}
}

func (cli *socketClient) recvResponseRoutine(conn io.Reader) {
	r := bufio.NewReader(conn)
	for {
		if !cli.IsRunning() {
			return
		}

		res := &types.Response{}
		err := types.ReadMessage(r, res)
		if err != nil {
			cli.stopForError(fmt.Errorf("read message: %w", err))
			return
		}

		switch r := res.Value.(type) {
		case *types.Response_Exception: // app responded with error
			// XXX After setting cli.err, release waiters (e.g. reqres.Done())
			cli.stopForError(errors.New(r.Exception.Error))
			return
		default:
			err := cli.didRecvResponse(res)
			if err != nil {
				cli.stopForError(err)
				return
			}
		}
	}
}

func (cli *socketClient) trackRequest(reqres *ReqRes) {
	// N.B. We must NOT hold the client state lock while checking this, or we
	// may deadlock with shutdown.
	if !cli.IsRunning() {
		return
	}

	cli.mtx.Lock()
	defer cli.mtx.Unlock()
	cli.reqSent.PushBack(reqres)
}

func (cli *socketClient) didRecvResponse(res *types.Response) error {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()

	// Get the first ReqRes.
	next := cli.reqSent.Front()
	if next == nil {
		return fmt.Errorf("unexpected response %T when no call was made", res.Value)
	}

	reqres := next.Value.(*ReqRes)
	if !resMatchesReq(reqres.Request, res) {
		return fmt.Errorf("unexpected response %T to the request %T", res.Value, reqres.Request.Value)
	}

	reqres.Response = res
	reqres.Done()            // release waiters
	cli.reqSent.Remove(next) // pop first item from linked list

	// Notify client listener if set (global callback).
	if cli.resCb != nil {
		cli.resCb(reqres.Request, res)
	}

	// Notify reqRes listener if set (request specific callback).
	//
	// NOTE: It is possible this callback isn't set on the reqres object. At this
	// point, in which case it will be called after, when it is set.
	reqres.InvokeCallback()

	return nil
}

//----------------------------------------

func (cli *socketClient) Flush(ctx context.Context) error {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestFlush())
	if err != nil {
		return err
	}
	reqRes.Wait()
	return nil
}

func (cli *socketClient) Echo(ctx context.Context, msg string) (*types.ResponseEcho, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestEcho(msg))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetEcho(), cli.Error()
}

func (cli *socketClient) Info(ctx context.Context, req *types.RequestInfo) (*types.ResponseInfo, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestInfo(req))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetInfo(), cli.Error()
}

func (cli *socketClient) Query(ctx context.Context, req *types.RequestQuery) (*types.ResponseQuery, error) {
	reqRes, err := cli.queueRequest(ctx, types.ToRequestQuery(req))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetQuery(), cli.Error()
}

func (cli *socketClient) ProcessDVSRequest(ctx context.Context, req *types.RequestProcessDVSRequest) (*types.ResponseProcessDVSRequest, error) {

	reqRes, err := cli.queueRequest(ctx, types.ToRequestProcessDvsRequest(req))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetProcessDvsRequest(), cli.Error()
}

func (cli *socketClient) ProcessDVSResponse(ctx context.Context, req *types.RequestProcessDVSResponse) (*types.ResponseProcessDVSResponse, error) {

	reqRes, err := cli.queueRequest(ctx, types.ToRequestProcessDvsResponse(req))
	if err != nil {
		return nil, err
	}
	if err := cli.Flush(ctx); err != nil {
		return nil, err
	}
	return reqRes.Response.GetProcessDvsResponse(), cli.Error()

}

// flushQueue marks as complete and discards all remaining pending requests
// from the queue.
func (cli *socketClient) flushQueue() {
	cli.mtx.Lock()
	defer cli.mtx.Unlock()

	// mark all in-flight messages as resolved (they will get cli.Error())
	for req := cli.reqSent.Front(); req != nil; req = req.Next() {
		reqres := req.Value.(*ReqRes)
		reqres.Done()
	}

	// mark all queued messages as resolved
LOOP:
	for {
		select {
		case reqres := <-cli.reqQueue:
			reqres.Done()
		default:
			break LOOP
		}
	}
}

// //----------------------------------------

func resMatchesReq(req *types.Request, res *types.Response) (ok bool) {
	switch req.Value.(type) {
	case *types.Request_Echo:
		_, ok = res.Value.(*types.Response_Echo)
	case *types.Request_Flush:
		_, ok = res.Value.(*types.Response_Flush)
	case *types.Request_Info:
		_, ok = res.Value.(*types.Response_Info)
	case *types.Request_ProcessDvsRequest:
		_, ok = res.Value.(*types.Response_ProcessDvsRequest)
	case *types.Request_ProcessDvsResponse:
		_, ok = res.Value.(*types.Response_ProcessDvsResponse)
	case *types.Request_Query:
		_, ok = res.Value.(*types.Response_Query)
	}
	return ok
}

func (cli *socketClient) stopForError(err error) {
	if !cli.IsRunning() {
		return
	}

	cli.mtx.Lock()
	if cli.err == nil {
		cli.err = err
	}
	cli.mtx.Unlock()

	cli.Logger.Error(fmt.Sprintf("Stopping avsi.socketClient for error: %v", err.Error()))
	if err := cli.Stop(); err != nil {
		cli.Logger.Error("Error stopping avsi.socketClient", "err", err)
	}
}

func (cli *socketClient) queueRequest(ctx context.Context, req *types.Request) (*ReqRes, error) {
	reqres := NewReqRes(req)

	// TODO: set cli.err if reqQueue times out
	select {
	case cli.reqQueue <- reqres:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Maybe auto-flush, or unset auto-flush
	switch req.Value.(type) {
	case *types.Request_Flush:
		cli.flushTimer.Unset()
	default:
		cli.flushTimer.Set()
	}

	return reqres, nil
}

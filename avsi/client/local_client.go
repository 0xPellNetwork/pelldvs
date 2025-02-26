package avsicli

import (
	"context"

	types "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/libs/service"
	cmtsync "github.com/0xPellNetwork/pelldvs/libs/sync"
)

// NOTE: use defer to unlock mutex because Application might panic (e.g., in
// case of malicious tx or query). It only makes sense for publicly exposed
// methods like CheckTx (/broadcast_tx_* RPC endpoint) or Query (/avsi_query
// RPC endpoint), but defers are used everywhere for the sake of consistency.
type localClient struct {
	service.BaseService

	mtx *cmtsync.Mutex
	types.Application
	Callback
}

var _ Client = (*localClient)(nil)

// NewLocalClient creates a local client, which wraps the application interface that
// Tendermint as the client will call to the application as the server. The only
// difference, is that the local client has a global mutex which enforces serialization
// of all the avsi calls from Tendermint to the Application.
func NewLocalClient(mtx *cmtsync.Mutex, app types.Application) Client {
	if mtx == nil {
		mtx = new(cmtsync.Mutex)
	}
	cli := &localClient{
		mtx:         mtx,
		Application: app,
	}
	cli.BaseService = *service.NewBaseService(nil, "localClient", cli)
	return cli
}

// -------------------------------------------------------
func (app *localClient) Error() error {
	return nil
}

func (app *localClient) Flush(context.Context) error {
	return nil
}

func (app *localClient) Echo(_ context.Context, msg string) (*types.ResponseEcho, error) {
	return &types.ResponseEcho{Message: msg}, nil
}

func (app *localClient) Info(ctx context.Context, req *types.RequestInfo) (*types.ResponseInfo, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.Info(ctx, req)
}

func (app *localClient) Query(ctx context.Context, req *types.RequestQuery) (*types.ResponseQuery, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.Query(ctx, req)
}

func (app *localClient) ProcessDVSRequest(ctx context.Context, req *types.RequestProcessDVSRequest) (*types.ResponseProcessDVSRequest, error) {

	app.mtx.Lock()
	defer app.mtx.Unlock()
	return app.Application.ProcessDVSRequest(ctx, req)
}

func (app *localClient) ProcessDVSResponse(ctx context.Context, req *types.RequestProcessDVSResponse) (*types.ResponseProcessDVSResponse, error) {

	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.Application.ProcessDVSResponse(ctx, req)
}

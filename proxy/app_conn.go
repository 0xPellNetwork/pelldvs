package proxy

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"

	avsicli "github.com/0xPellNetwork/pelldvs/avsi/client"
	"github.com/0xPellNetwork/pelldvs/avsi/types"
)

//----------------------------------------------------------------------------------------
// Enforce which avsi msgs can be sent on a connection at the type level

type AppConnDvs interface {
	Error() error

	ProcessDVSRequest(context.Context, *types.RequestProcessDVSRequest) (*types.ResponseProcessDVSRequest, error)
	ProcessDVSResponse(context.Context, *types.RequestProcessDVSResponse) (*types.ResponseProcessDVSResponse, error)
}

type AppConnQuery interface {
	Error() error

	Echo(context.Context, string) (*types.ResponseEcho, error)
	Info(context.Context, *types.RequestInfo) (*types.ResponseInfo, error)
	Query(context.Context, *types.RequestQuery) (*types.ResponseQuery, error)
}

//-----------------------------------------------------------------------------------------
// Implements AppConnDvs (subset of avsicli.Client)

type appConnDvs struct {
	metrics *Metrics
	appConn avsicli.Client
}

var _ AppConnDvs = (*appConnDvs)(nil)

func NewAppConnDvs(appConn avsicli.Client, metrics *Metrics) *appConnDvs {
	return &appConnDvs{
		metrics: metrics,
		appConn: appConn,
	}
}

func (app *appConnDvs) Error() error {
	return app.appConn.Error()
}

func (app *appConnDvs) ProcessDVSRequest(ctx context.Context, req *types.RequestProcessDVSRequest) (*types.ResponseProcessDVSRequest, error) {

	defer addTimeSample(app.metrics.MethodTimingSeconds.With("method", "processrequest", "type", "sync"))()
	return app.appConn.ProcessDVSRequest(ctx, req)
}

func (app *appConnDvs) ProcessDVSResponse(ctx context.Context, req *types.RequestProcessDVSResponse) (*types.ResponseProcessDVSResponse, error) {

	defer addTimeSample(app.metrics.MethodTimingSeconds.With("method", "PostRequest", "type", "sync"))()
	return app.appConn.ProcessDVSResponse(ctx, req)

}

//------------------------------------------------
// Implements AppConnQuery (subset of avsicli.Client)

type appConnQuery struct {
	metrics *Metrics
	appConn avsicli.Client
}

func NewAppConnQuery(appConn avsicli.Client, metrics *Metrics) AppConnQuery {
	return &appConnQuery{
		metrics: metrics,
		appConn: appConn,
	}
}

func (app *appConnQuery) Error() error {
	return app.appConn.Error()
}

func (app *appConnQuery) Echo(ctx context.Context, msg string) (*types.ResponseEcho, error) {
	defer addTimeSample(app.metrics.MethodTimingSeconds.With("method", "echo", "type", "sync"))()
	return app.appConn.Echo(ctx, msg)
}

func (app *appConnQuery) Info(ctx context.Context, req *types.RequestInfo) (*types.ResponseInfo, error) {
	defer addTimeSample(app.metrics.MethodTimingSeconds.With("method", "info", "type", "sync"))()
	return app.appConn.Info(ctx, req)
}

func (app *appConnQuery) Query(ctx context.Context, req *types.RequestQuery) (*types.ResponseQuery, error) {
	defer addTimeSample(app.metrics.MethodTimingSeconds.With("method", "query", "type", "sync"))()
	return app.appConn.Query(ctx, req)
}

// addTimeSample returns a function that, when called, adds an observation to m.
// The observation added to m is the number of seconds ellapsed since addTimeSample
// was initially called. addTimeSample is meant to be called in a defer to calculate
// the amount of time a function takes to complete.
func addTimeSample(m metrics.Histogram) func() {
	start := time.Now()
	return func() { m.Observe(time.Since(start).Seconds()) }
}

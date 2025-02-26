package types

import (
	"context"
)

// Application is an interface that enables any finite, deterministic state machine
// to be driven by a blockchain-based replication engine via the ABCI.
type Application interface {
	// Info/Query Connection
	Info(context.Context, *RequestInfo) (*ResponseInfo, error)    // Return application info
	Query(context.Context, *RequestQuery) (*ResponseQuery, error) // Query for state

	//dvs connection
	ProcessDVSRequest(context.Context, *RequestProcessDVSRequest) (*ResponseProcessDVSRequest, error)
	ProcessDVSResponse(context.Context, *RequestProcessDVSResponse) (*ResponseProcessDVSResponse, error)
}

//-------------------------------------------------------
// BaseApplication is a base form of Application

var _ Application = (*BaseApplication)(nil)

type BaseApplication struct{}

func NewBaseApplication() *BaseApplication {
	return &BaseApplication{}
}

func (BaseApplication) Info(context.Context, *RequestInfo) (*ResponseInfo, error) {
	return &ResponseInfo{}, nil
}

func (BaseApplication) Query(context.Context, *RequestQuery) (*ResponseQuery, error) {
	return &ResponseQuery{Code: CodeTypeOK}, nil
}

func (BaseApplication) ProcessDVSRequest(context.Context, *RequestProcessDVSRequest) (*ResponseProcessDVSRequest, error) {
	return &ResponseProcessDVSRequest{}, nil
}

func (BaseApplication) ProcessDVSResponse(context.Context, *RequestProcessDVSResponse) (*ResponseProcessDVSResponse, error) {
	return &ResponseProcessDVSResponse{}, nil
}

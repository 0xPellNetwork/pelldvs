package coregrpc

import (
	"context"

	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"

	core "github.com/0xPellNetwork/pelldvs/rpc/core"
)

type DVSRequestAPIServerAPI struct {
	env *core.Environment
}

func (api *DVSRequestAPIServerAPI) Ping(ctx context.Context, req *RequestPing) (*ResponsePing, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}

func (api *DVSRequestAPIServerAPI) RequestDvsSync(ctx context.Context, req *DVSRequest) (*ResultDvsRequestCommit, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestDvsSync not implemented")
}

func (api *DVSRequestAPIServerAPI) RequestDvsAsync(ctx context.Context, req *DVSRequest) (*ResultRequestDvsAsync, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestDvsAsync not implemented")
}

func (api *DVSRequestAPIServerAPI) QueryDvsRequest(ctx context.Context, req *QueryDvsRequestParam) (*ResultDvsRequestCommit, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryDvsRequest not implemented")
}

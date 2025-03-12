package core

import (
	"context"

	avsi "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/libs/bytes"
	"github.com/0xPellNetwork/pelldvs/proxy"
	ctypes "github.com/0xPellNetwork/pelldvs/rpc/core/types"
	rpctypes "github.com/0xPellNetwork/pelldvs/rpc/jsonrpc/types"
)

// AVSIQuery queries the application for some information.
func (env *Environment) AVSIQuery(
	_ *rpctypes.Context,
	path string,
	data bytes.HexBytes,
	height int64,
	prove bool,
) (*ctypes.ResultAVSIQuery, error) {
	resQuery, err := env.ProxyAppQuery.Query(context.TODO(), &avsi.RequestQuery{
		Path:   path,
		Data:   data,
		Height: height,
		Prove:  prove,
	})
	if err != nil {
		return nil, err
	}

	return &ctypes.ResultAVSIQuery{Response: *resQuery}, nil
}

// AVSIInfo gets some info about the application.
func (env *Environment) AVSIInfo(_ *rpctypes.Context) (*ctypes.ResultAVSIInfo, error) {
	resInfo, err := env.ProxyAppQuery.Info(context.TODO(), proxy.RequestInfo)
	if err != nil {
		return nil, err
	}

	return &ctypes.ResultAVSIInfo{Response: *resInfo}, nil
}

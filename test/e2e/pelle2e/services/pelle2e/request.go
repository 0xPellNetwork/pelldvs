package pelle2e

import (
	"context"
	"fmt"

	"github.com/labstack/gommon/random"

	"github.com/0xPellNetwork/pelldvs/avsi/types"
	avsitypes "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/rpc/client/http"
	ctypes "github.com/0xPellNetwork/pelldvs/rpc/core/types"
)

func (per *PellDVSE2ERunner) PrepareRequest(
	ctx context.Context,
	groupNumbers []uint32,
	groupThresholdPercentages []uint32,
) (*E2EContext, *types.DVSRequest, error) {

	blockNumber, err := per.Client.BlockNumber(ctx)
	if err != nil {
		return nil, nil, err
	}
	chainID, err := per.Client.ChainID(ctx)
	if err != nil {
		return nil, nil, err
	}

	key := random.String(10)
	value := random.String(10)

	kvc := NewKVStoreAppContext(key, value)
	eectx := NewE2EContext(kvc)

	data := []byte(fmt.Sprintf("%s=%s", eectx.KVStoreApp.Key, eectx.KVStoreApp.Value))

	per.logger.Info("request data", "key", key, "value", value, "data", data)

	req := &avsitypes.DVSRequest{
		Data:                      data,
		Height:                    int64(blockNumber),
		ChainId:                   chainID.Int64(),
		GroupNumbers:              groupNumbers,
		GroupThresholdPercentages: groupThresholdPercentages,
	}

	return eectx, req, nil
}

func (per *PellDVSE2ERunner) RequestDVSAsync(ctx context.Context, req *types.DVSRequest) (*ctypes.ResultRequestDvsAsync, error) {
	httpClient, err := http.New(per.DVSNodeRPCURL, "")
	if err != nil {
		return nil, err
	}

	result, err := httpClient.RequestDVSAsync(
		ctx,
		req.Data,
		req.Height,
		req.ChainId,
		req.GroupNumbers,
		req.GroupThresholdPercentages,
	)
	return result, err
}

func (per *PellDVSE2ERunner) QueryRequest(ctx context.Context, hash string) (*ctypes.ResultDvsRequest, error) {
	httpClient, err := http.New(per.DVSNodeRPCURL, "")
	if err != nil {
		return nil, err
	}
	result, err := httpClient.QueryRequest(ctx, hash)
	return result, err
}

func (per *PellDVSE2ERunner) SearchRequest(ctx context.Context, query string, pagePtr, perPagePtr *int) (*ctypes.ResultDvsRequestSearch, error) {
	httpClient, err := http.New(per.DVSNodeRPCURL, "")
	if err != nil {
		return nil, err
	}
	result, err := httpClient.SearchRequest(ctx, query, pagePtr, perPagePtr)
	return result, err
}

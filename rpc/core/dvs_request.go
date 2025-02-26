package core

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	avsiTypes "github.com/0xPellNetwork/pelldvs/avsi/types"
	cmtmath "github.com/0xPellNetwork/pelldvs/libs/math"
	cmtquery "github.com/0xPellNetwork/pelldvs/libs/query"
	ctypes "github.com/0xPellNetwork/pelldvs/rpc/core/types"
	rpctypes "github.com/0xPellNetwork/pelldvs/rpc/jsonrpc/types"
	"github.com/0xPellNetwork/pelldvs/state/requestindex/null"
	"github.com/0xPellNetwork/pelldvs/types"
)

func (env *Environment) RequestDVS(ctx *rpctypes.Context,
	data []byte,
	height int64,
	chainid int64,
	groupNumbers []uint32,
	groupThresholdPercentages []uint32,
) (*ctypes.ResultRequest, error) {

	req := avsiTypes.DVSRequest{
		Data:                      data,
		Height:                    height,
		ChainId:                   chainid,
		GroupNumbers:              groupNumbers,
		GroupThresholdPercentages: groupThresholdPercentages,
	}

	_, err := env.DVSReactor.OnRequest(req)

	if err != nil {
		return &ctypes.ResultRequest{}, err
	}

	return &ctypes.ResultRequest{}, nil
}

func (env *Environment) RequestDVSAsync(ctx *rpctypes.Context,
	data []byte,
	height int64,
	chainid int64,
	groupNumbers []uint32,
	groupThresholdPercentages []uint32,
) (*ctypes.ResultRequestDvsAsync, error) {

	req := avsiTypes.DVSRequest{
		Data:                      data,
		Height:                    height,
		ChainId:                   chainid,
		GroupNumbers:              groupNumbers,
		GroupThresholdPercentages: groupThresholdPercentages,
	}

	rawBytesDvsRequest, err := proto.Marshal(&req)
	if err != nil {
		return nil, err
	}
	hash := types.DvsRequest(rawBytesDvsRequest).Hash()

	go func() {
		if _, err := env.DVSReactor.OnRequest(req); err != nil {
			env.Logger.Error("RequestDvsAsync", "module", "rpc", "func", "OnRequest", "err", err)
		}
	}()

	return &ctypes.ResultRequestDvsAsync{Hash: hash}, nil

}

func (env *Environment) QueryRequest(_ *rpctypes.Context, hash string) (*ctypes.ResultDvsRequest, error) {

	hashAsBytes, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	// if index is disabled, return error
	if _, ok := env.DvsRequestIndexer.(*null.DvsRequestIndex); ok {
		return nil, fmt.Errorf("dvs indexing is disabled")
	}

	r, err := env.DvsRequestIndexer.Get(hashAsBytes)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, fmt.Errorf("dvs (%X) not found", hash)
	}

	return &ctypes.ResultDvsRequest{
		DvsRequest:                 r.DvsRequest,
		DvsResponse:                r.DvsResponse,
		ResponseProcessDvsRequest:  r.ResponseProcessDvsRequest,
		ResponseProcessDVSResponse: r.ResponseProcessDvsResponse,
	}, nil
}

func (env *Environment) SearchRequest(
	ctx *rpctypes.Context,
	query string,
	prove bool,
	pagePtr, perPagePtr *int,
) (*ctypes.ResultDvsRequestSearch, error) {

	// if index is disabled, return error
	if _, ok := env.DvsRequestIndexer.(*null.DvsRequestIndex); ok {
		return nil, errors.New("dvs request indexing is disabled")
	} else if len(query) > maxQueryLength {
		return nil, errors.New("maximum query length exceeded")
	}

	q, err := cmtquery.New(query)
	if err != nil {
		return nil, err
	}

	results, err := env.DvsRequestIndexer.Search(ctx.Context(), q)
	if err != nil {
		return nil, err
	}

	// paginate results
	totalCount := len(results)
	perPage := env.validatePerPage(perPagePtr)

	page, err := validatePage(pagePtr, perPage, totalCount)
	if err != nil {
		return nil, err
	}

	skipCount := validateSkipCount(page, perPage)
	pageSize := cmtmath.MinInt(perPage, totalCount-skipCount)

	apiResults := make([]*ctypes.ResultDvsRequest, 0, pageSize)
	for i := skipCount; i < skipCount+pageSize; i++ {
		r := results[i]

		rawBytesDvsRequest, err := proto.Marshal(r.DvsRequest)
		if err != nil {
			return nil, err
		}
		hash := types.DvsRequest(rawBytesDvsRequest).Hash()

		apiResults = append(apiResults, &ctypes.ResultDvsRequest{
			DvsRequest:                 r.DvsRequest,
			DvsResponse:                r.DvsResponse,
			ResponseProcessDvsRequest:  r.ResponseProcessDvsRequest,
			ResponseProcessDVSResponse: r.ResponseProcessDvsResponse,
			Hash:                       hash,
		})
	}

	return &ctypes.ResultDvsRequestSearch{DvsRequests: apiResults, TotalCount: totalCount}, nil

}

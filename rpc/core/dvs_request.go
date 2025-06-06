package core

import (
	"encoding/hex"
	"errors"
	"fmt"

	avsiTypes "github.com/0xPellNetwork/pelldvs/avsi/types"
	"github.com/0xPellNetwork/pelldvs/libs/bytes"
	cmtmath "github.com/0xPellNetwork/pelldvs/libs/math"
	cmtquery "github.com/0xPellNetwork/pelldvs/libs/query"
	ctypes "github.com/0xPellNetwork/pelldvs/rpc/core/types"
	rpctypes "github.com/0xPellNetwork/pelldvs/rpc/jsonrpc/types"
	"github.com/0xPellNetwork/pelldvs/state/requestindex/null"
)

func (env *Environment) RequestDVS(ctx *rpctypes.Context,
	data []byte,
	height int64,
	chainid int64,
	groupNumbers []uint32,
	groupThresholdPercentages []uint32,
) (*ctypes.ResultRequest, error) {
	request := avsiTypes.DVSRequest{
		Data:                      data,
		Height:                    height,
		ChainId:                   chainid,
		GroupNumbers:              groupNumbers,
		GroupThresholdPercentages: groupThresholdPercentages,
	}

	err := env.DVSReactor.HandleDVSRequest(request)
	if err != nil {
		return &ctypes.ResultRequest{}, err
	}

	return &ctypes.ResultRequest{}, nil
}

func (env *Environment) RequestDVSAsync(ctx *rpctypes.Context,
	data []byte,
	height int64,
	chainID int64,
	groupNumbers []uint32,
	groupThresholdPercentages []uint32,
) (*ctypes.ResultRequestDvsAsync, error) {
	request := avsiTypes.DVSRequest{
		Data:                      data,
		Height:                    height,
		ChainId:                   chainID,
		GroupNumbers:              groupNumbers,
		GroupThresholdPercentages: groupThresholdPercentages,
	}
	go func() {
		if err := env.DVSReactor.HandleDVSRequest(request); err != nil {
			env.Logger.Error("RequestDvsAsync", "module", "rpc", "func", "HandleDVSRequest", "err", err)
		}
	}()

	return &ctypes.ResultRequestDvsAsync{
		Hash: bytes.HexBytes(request.Hash()),
	}, nil
}

// QueryRequest allows you to query for a DVS request result. It returns a
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

// SearchRequest allows you to query for multiple DVS request results. It returns a
func (env *Environment) SearchRequest(
	ctx *rpctypes.Context,
	query string,
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
		hash := []byte(r.DvsRequest.Hash())

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

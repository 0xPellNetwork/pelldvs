package core

import (
	ctypes "github.com/0xPellNetwork/pelldvs/rpc/core/types"
	rpctypes "github.com/0xPellNetwork/pelldvs/rpc/jsonrpc/types"
)

// Health gets node health. Returns empty result (200 OK) on success, no
// response - in case of an error.
func (env *Environment) Health(*rpctypes.Context) (*ctypes.ResultHealth, error) {
	return &ctypes.ResultHealth{}, nil
}

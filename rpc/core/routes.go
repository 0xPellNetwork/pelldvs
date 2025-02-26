package core

import (
	rpc "github.com/0xPellNetwork/pelldvs/rpc/jsonrpc/server"
)

// TODO: better system than "unsafe" prefix

type RoutesMap map[string]*rpc.RPCFunc

// Routes is a map of available routes.
func (env *Environment) GetRoutes() RoutesMap {
	return RoutesMap{
		// info API
		"health":   rpc.NewRPCFunc(env.Health, ""),
		"net_info": rpc.NewRPCFunc(env.NetInfo, ""),

		// status API
		"avsi_query": rpc.NewRPCFunc(env.AVSIQuery, "path,data,height,prove"),
		"avsi_info":  rpc.NewRPCFunc(env.AVSIInfo, "", rpc.Cacheable()),

		// dvs API
		"request_dvs":       rpc.NewRPCFunc(env.RequestDVS, "data,height,chainid,group_numbers,group_threshold_percentages"),
		"request_dvs_async": rpc.NewRPCFunc(env.RequestDVSAsync, "data,height,chainid,group_numbers,group_threshold_percentages"),
		"query_request":     rpc.NewRPCFunc(env.QueryRequest, "hash"),
		"search_request":    rpc.NewRPCFunc(env.SearchRequest, "query,page,per_page,order_by"),
	}
}

// AddUnsafeRoutes adds unsafe routes.
func (env *Environment) AddUnsafeRoutes(routes RoutesMap) {
	// control API
	routes["dial_seeds"] = rpc.NewRPCFunc(env.UnsafeDialSeeds, "seeds")
	routes["dial_peers"] = rpc.NewRPCFunc(env.UnsafeDialPeers, "peers,persistent,unconditional,private")
}

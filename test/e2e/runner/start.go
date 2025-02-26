package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/0xPellNetwork/pelldvs-libs/log"
	e2e "github.com/0xPellNetwork/pelldvs/test/e2e/pkg"
	"github.com/0xPellNetwork/pelldvs/test/e2e/pkg/infra"
)

func Start(ctx context.Context, testnet *e2e.Testnet, p infra.Provider) error {
	if len(testnet.Nodes) == 0 {
		return fmt.Errorf("no nodes in testnet")
	}

	// Nodes are already sorted by name. Sort them by name then startAt,
	// which gives the overall order startAt, mode, name.
	nodeQueue := testnet.Nodes
	sort.SliceStable(nodeQueue, func(i, j int) bool {
		a, b := nodeQueue[i], nodeQueue[j]
		switch {
		case a.Mode == b.Mode:
			return false
		case a.Mode == e2e.ModeSeed:
			return true
		case a.Mode == e2e.ModeValidator && b.Mode == e2e.ModeFull:
			return true
		}
		return false
	})

	sort.SliceStable(nodeQueue, func(i, j int) bool {
		return nodeQueue[i].StartAt < nodeQueue[j].StartAt
	})

	if nodeQueue[0].StartAt > 0 {
		return fmt.Errorf("no initial nodes in testnet")
	}

	// Start initial nodes (StartAt: 0)
	logger.Info("Starting initial network nodes...")
	nodesAtZero := make([]*e2e.Node, 0)
	for len(nodeQueue) > 0 && nodeQueue[0].StartAt == 0 {
		nodesAtZero = append(nodesAtZero, nodeQueue[0])
		nodeQueue = nodeQueue[1:]
	}
	err := p.StartNodes(context.Background(), nodesAtZero...)
	if err != nil {
		return err
	}
	for _, node := range nodesAtZero {
		if node.PrometheusProxyPort > 0 {
			logger.Info("start", "msg",
				log.NewLazySprintf("Node %v up on http://%s:%v; with Prometheus on http://%s:%v/metrics",
					node.Name,
					node.ExternalIP,
					node.ProxyPort,
					node.ExternalIP,
					node.PrometheusProxyPort,
				),
			)
		} else {
			logger.Info("start", "msg", log.NewLazySprintf("Node %v up on http://%s:%v",
				node.Name,
				node.ExternalIP,
				node.ProxyPort,
			))
		}
	}

	networkHeight := testnet.InitialHeight

	// Wait for initial height
	logger.Info("Waiting for initial height",
		"height", networkHeight,
		"nodes", len(testnet.Nodes)-len(nodeQueue),
		"pending", len(nodeQueue))

	for _, node := range nodeQueue {
		if node.StartAt > networkHeight {
			// if we're starting a node that's ahead of
			// the last known height of the network, then
			// we should make sure that the rest of the
			// network has reached at least the height
			// that this node will start at before we
			// start the node.

			networkHeight = node.StartAt

			logger.Info("Waiting for network to advance before starting catch up node",
				"node", node.Name,
				"height", networkHeight)

		}

		logger.Info("Starting catch up node", "node", node.Name, "height", node.StartAt)

		err := p.StartNodes(context.Background(), node)
		if err != nil {
			return err
		}

	}

	return nil
}

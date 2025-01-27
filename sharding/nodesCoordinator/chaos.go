package nodesCoordinator

import (
	"github.com/multiversx/mx-chain-go/chaos"
	"github.com/multiversx/mx-chain-go/chaosAdapters"
)

func learnChaosContext(eligibleNodesMap map[uint32][]Validator, waitingNodesMap map[uint32][]Validator) {
	eligibleNodes := make([]chaosAdapters.Validator, 0)
	waitingNodes := make([]chaosAdapters.Validator, 0)

	for _, v := range eligibleNodesMap {
		for _, item := range v {
			eligibleNodes = append(eligibleNodes, item)
		}
	}

	for _, v := range waitingNodesMap {
		for _, item := range v {
			waitingNodes = append(waitingNodes, item)
		}
	}

	chaos.Seed(eligibleNodes, waitingNodes)
}

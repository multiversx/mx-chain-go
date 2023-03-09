package mock

import (
	"sort"

	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

type nodesHandlerMock struct {
	allNodes []nodesCoordinator.GenesisNodeInfoHandler
}

// NewNodesHandlerMock -
func NewNodesHandlerMock(
	initialNodesSetup genesis.InitialNodesHandler,
) (*nodesHandlerMock, error) {

	eligible, waiting := initialNodesSetup.InitialNodesInfo()

	allNodes := make([]nodesCoordinator.GenesisNodeInfoHandler, 0)
	keys := make([]uint32, 0)
	for shard := range eligible {
		keys = append(keys, shard)
	}

	//it is important that the processing is done in a deterministic way
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, shardID := range keys {
		allNodes = append(allNodes, eligible[shardID]...)
		allNodes = append(allNodes, waiting[shardID]...)
	}

	return &nodesHandlerMock{
		allNodes: allNodes,
	}, nil
}

// GetAllNodes -
func (nhm *nodesHandlerMock) GetAllNodes() []nodesCoordinator.GenesisNodeInfoHandler {
	stakedNodes := make([]nodesCoordinator.GenesisNodeInfoHandler, 0)
	stakedNodes = append(stakedNodes, nhm.allNodes...)

	return stakedNodes
}

// GetDelegatedNodes -
func (nhm *nodesHandlerMock) GetDelegatedNodes(_ []byte) []nodesCoordinator.GenesisNodeInfoHandler {
	return make([]nodesCoordinator.GenesisNodeInfoHandler, 0)
}

// IsInterfaceNil returns if underlying object is true
func (nhm *nodesHandlerMock) IsInterfaceNil() bool {
	return nhm == nil
}

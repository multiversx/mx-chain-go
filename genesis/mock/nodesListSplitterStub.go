package mock

import "github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"

// NodesListSplitterStub -
type NodesListSplitterStub struct {
	GetAllNodesCalled       func() []nodesCoordinator.GenesisNodeInfoHandler
	GetDelegatedNodesCalled func(delegationScAddress []byte) []nodesCoordinator.GenesisNodeInfoHandler
}

// GetAllNodes -
func (nlss *NodesListSplitterStub) GetAllNodes() []nodesCoordinator.GenesisNodeInfoHandler {
	if nlss.GetAllNodesCalled != nil {
		return nlss.GetAllNodesCalled()
	}

	return make([]nodesCoordinator.GenesisNodeInfoHandler, 0)
}

// GetDelegatedNodes -
func (nlss *NodesListSplitterStub) GetDelegatedNodes(delegationScAddress []byte) []nodesCoordinator.GenesisNodeInfoHandler {
	if nlss.GetDelegatedNodesCalled != nil {
		return nlss.GetDelegatedNodesCalled(delegationScAddress)
	}

	return make([]nodesCoordinator.GenesisNodeInfoHandler, 0)
}

// IsInterfaceNil -
func (nlss *NodesListSplitterStub) IsInterfaceNil() bool {
	return nlss == nil
}

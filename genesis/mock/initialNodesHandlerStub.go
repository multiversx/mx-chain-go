package mock

import "github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"

// InitialNodesHandlerStub -
type InitialNodesHandlerStub struct {
	InitialNodesInfoCalled               func() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
	MinNumberOfNodesCalled               func() uint32
	MinNumberOfNodesWithHysteresisCalled func() uint32
}

// InitialNodesInfo -
func (inhs *InitialNodesHandlerStub) InitialNodesInfo() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
	if inhs.InitialNodesInfoCalled != nil {
		return inhs.InitialNodesInfoCalled()
	}

	return make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler), make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
}

// MinNumberOfNodes -
func (inhs *InitialNodesHandlerStub) MinNumberOfNodes() uint32 {
	if inhs.MinNumberOfNodesCalled != nil {
		return inhs.MinNumberOfNodesCalled()
	}

	return 0
}

// MinNumberOfNodesWithHysteresis -
func (inhs *InitialNodesHandlerStub) MinNumberOfNodesWithHysteresis() uint32 {
	if inhs.MinNumberOfNodesWithHysteresisCalled != nil {
		return inhs.MinNumberOfNodesWithHysteresisCalled()
	}

	return inhs.MinNumberOfNodes()
}

// IsInterfaceNil -
func (inhs *InitialNodesHandlerStub) IsInterfaceNil() bool {
	return inhs == nil
}

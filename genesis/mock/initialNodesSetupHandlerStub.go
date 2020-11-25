package mock

import "github.com/ElrondNetwork/elrond-go/sharding"

// InitialNodesSetupHandlerStub -
type InitialNodesSetupHandlerStub struct {
	InitialNodesInfoCalled               func() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler)
	MinNumberOfNodesCalled               func() uint32
	MinNumberOfNodesWithHysteresisCalled func() uint32
}

// InitialNodesInfo -
func (inshs *InitialNodesSetupHandlerStub) InitialNodesInfo() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler) {
	if inshs.InitialNodesInfoCalled != nil {
		return inshs.InitialNodesInfoCalled()
	}

	return make(map[uint32][]sharding.GenesisNodeInfoHandler), make(map[uint32][]sharding.GenesisNodeInfoHandler)
}

// MinNumberOfNodes -
func (inshs *InitialNodesSetupHandlerStub) MinNumberOfNodes() uint32 {
	if inshs.MinNumberOfNodesCalled != nil {
		return inshs.MinNumberOfNodesCalled()
	}

	return 1
}

// MinNumberOfNodesWithHysteresis -
func (inshs *InitialNodesSetupHandlerStub) MinNumberOfNodesWithHysteresis() uint32 {
	if inshs.MinNumberOfNodesWithHysteresisCalled != nil {
		return inshs.MinNumberOfNodesWithHysteresisCalled()
	}

	return inshs.MinNumberOfNodes()
}

// IsInterfaceNil -
func (inshs *InitialNodesSetupHandlerStub) IsInterfaceNil() bool {
	return inshs == nil
}

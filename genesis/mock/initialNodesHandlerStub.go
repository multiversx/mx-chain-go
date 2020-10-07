package mock

import "github.com/ElrondNetwork/elrond-go/sharding"

// InitialNodesHandlerStub -
type InitialNodesHandlerStub struct {
	InitialNodesInfoCalled        func() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler)
	MinNumberOfNodesCalled        func() uint32
	MinShardHysteresisNodesCalled func() uint32
	MinMetaHysteresisNodesCalled  func() uint32
}

// InitialNodesInfo -
func (inhs *InitialNodesHandlerStub) InitialNodesInfo() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler) {
	if inhs.InitialNodesInfoCalled != nil {
		return inhs.InitialNodesInfoCalled()
	}

	return make(map[uint32][]sharding.GenesisNodeInfoHandler), make(map[uint32][]sharding.GenesisNodeInfoHandler)
}

// MinNumberOfNodes -
func (inhs *InitialNodesHandlerStub) MinNumberOfNodes() uint32 {
	if inhs.MinNumberOfNodesCalled != nil {
		return inhs.MinNumberOfNodesCalled()
	}

	return 0
}

// MinShardHysteresisNodes -
func (inhs *InitialNodesHandlerStub) MinShardHysteresisNodes() uint32 {
	if inhs.MinShardHysteresisNodesCalled != nil {
		return inhs.MinShardHysteresisNodesCalled()
	}
	return 0
}

// MinMetaHysteresisNodes -
func (inhs *InitialNodesHandlerStub) MinMetaHysteresisNodes() uint32 {
	if inhs.MinMetaHysteresisNodesCalled != nil {
		return inhs.MinMetaHysteresisNodesCalled()
	}
	return 0
}

// IsInterfaceNil -
func (inhs *InitialNodesHandlerStub) IsInterfaceNil() bool {
	return inhs == nil
}

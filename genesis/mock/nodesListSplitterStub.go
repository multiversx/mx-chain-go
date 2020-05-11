package mock

import "github.com/ElrondNetwork/elrond-go/sharding"

// NodesListSplitterStub -
type NodesListSplitterStub struct {
	GetAllNodesCalled       func() []sharding.GenesisNodeInfoHandler
	GetDelegatedNodesCalled func(delegationScAddress []byte) []sharding.GenesisNodeInfoHandler
}

// GetAllNodes -
func (nlss *NodesListSplitterStub) GetAllNodes() []sharding.GenesisNodeInfoHandler {
	if nlss.GetAllNodesCalled != nil {
		return nlss.GetAllNodesCalled()
	}

	return make([]sharding.GenesisNodeInfoHandler, 0)
}

// GetDelegatedNodes -
func (nlss *NodesListSplitterStub) GetDelegatedNodes(delegationScAddress []byte) []sharding.GenesisNodeInfoHandler {
	if nlss.GetDelegatedNodesCalled != nil {
		return nlss.GetDelegatedNodesCalled(delegationScAddress)
	}

	return make([]sharding.GenesisNodeInfoHandler, 0)
}

// IsInterfaceNil -
func (nlss *NodesListSplitterStub) IsInterfaceNil() bool {
	return nlss == nil
}

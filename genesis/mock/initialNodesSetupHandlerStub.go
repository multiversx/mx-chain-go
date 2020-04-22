package mock

import "github.com/ElrondNetwork/elrond-go/sharding"

// InitialNodesSetupHandlerStub -
type InitialNodesSetupHandlerStub struct {
	InitialNodesInfoCalled func() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler)
}

// InitialNodesInfo -
func (inshs *InitialNodesSetupHandlerStub) InitialNodesInfo() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler) {
	if inshs.InitialNodesInfoCalled != nil {
		return inshs.InitialNodesInfoCalled()
	}

	return make(map[uint32][]sharding.GenesisNodeInfoHandler), make(map[uint32][]sharding.GenesisNodeInfoHandler)
}

// IsInterfaceNil -
func (inshs *InitialNodesSetupHandlerStub) IsInterfaceNil() bool {
	return inshs == nil
}

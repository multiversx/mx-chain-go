package mock

import (
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/update/mock"
)

// GenesisNodesSetupFactoryMock -
type GenesisNodesSetupFactoryMock struct {
	CreateNodesSetupCalled func(args *sharding.NodesSetupArgs) (sharding.GenesisNodesSetupHandler, error)
}

// CreateNodesSetup -
func (f *GenesisNodesSetupFactoryMock) CreateNodesSetup(args *sharding.NodesSetupArgs) (sharding.GenesisNodesSetupHandler, error) {
	if f.CreateNodesSetupCalled != nil {
		return f.CreateNodesSetupCalled(args)
	}
	return &mock.GenesisNodesSetupHandlerStub{}, nil
}

// IsInterfaceNil -
func (f *GenesisNodesSetupFactoryMock) IsInterfaceNil() bool {
	return f == nil
}

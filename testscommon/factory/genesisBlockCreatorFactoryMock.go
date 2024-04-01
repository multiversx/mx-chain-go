package factory

import (
	"github.com/multiversx/mx-chain-go/genesis/process"
)

// GenesisBlockCreatorFactoryMock -
type GenesisBlockCreatorFactoryMock struct {
	CreateGenesisBlockCreatorCalled func(args process.ArgsGenesisBlockCreator) (process.GenesisBlockCreatorHandler, error)
}

// CreateGenesisBlockCreator -
func (gbf *GenesisBlockCreatorFactoryMock) CreateGenesisBlockCreator(args process.ArgsGenesisBlockCreator) (process.GenesisBlockCreatorHandler, error) {
	if gbf.CreateGenesisBlockCreatorCalled != nil {
		return gbf.CreateGenesisBlockCreatorCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (gbf *GenesisBlockCreatorFactoryMock) IsInterfaceNil() bool {
	return gbf == nil
}

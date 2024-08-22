package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory/shard/data"
)

// PreProcessorContainerFactoryCreatorMock -
type PreProcessorContainerFactoryCreatorMock struct {
	CreatePreProcessorContainerFactoryCalled func(args data.ArgPreProcessorsContainerFactory) (process.PreProcessorsContainerFactory, error)
}

// CreatePreProcessorContainerFactory -
func (f *PreProcessorContainerFactoryCreatorMock) CreatePreProcessorContainerFactory(
	args data.ArgPreProcessorsContainerFactory,
) (process.PreProcessorsContainerFactory, error) {
	if f.CreatePreProcessorContainerFactoryCalled != nil {
		return f.CreatePreProcessorContainerFactoryCalled(args)
	}

	return &PreProcessorsContainerFactoryMock{}, nil
}

// IsInterfaceNil -
func (f *PreProcessorContainerFactoryCreatorMock) IsInterfaceNil() bool {
	return f == nil
}

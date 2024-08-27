package sovereign

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory/shard/data"
)

type sovereignPreProcessorContainerFactoryCreator struct {
}

// NewSovereignPreProcessorContainerFactoryCreator creates a pre-processors container factory creator
func NewSovereignPreProcessorContainerFactoryCreator() *sovereignPreProcessorContainerFactoryCreator {
	return &sovereignPreProcessorContainerFactoryCreator{}
}

// CreatePreProcessorContainerFactory creates a pre-processor container factory for shard normal run type
func (f *sovereignPreProcessorContainerFactoryCreator) CreatePreProcessorContainerFactory(
	args data.ArgPreProcessorsContainerFactory,
) (process.PreProcessorsContainerFactory, error) {
	return NewSovereignPreProcessorsContainerFactory(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *sovereignPreProcessorContainerFactoryCreator) IsInterfaceNil() bool {
	return f == nil
}

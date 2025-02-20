package shard

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory/shard/data"
)

type preProcessorContainerFactoryCreator struct {
}

// NewPreProcessorContainerFactoryCreator creates a sovereign pre-processors container factory creator
func NewPreProcessorContainerFactoryCreator() *preProcessorContainerFactoryCreator {
	return &preProcessorContainerFactoryCreator{}
}

// CreatePreProcessorContainerFactory creates a pre-processor container factory for sovereign run type
func (f *preProcessorContainerFactoryCreator) CreatePreProcessorContainerFactory(
	args data.ArgPreProcessorsContainerFactory,
) (process.PreProcessorsContainerFactory, error) {
	return NewPreProcessorsContainerFactory(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *preProcessorContainerFactoryCreator) IsInterfaceNil() bool {
	return f == nil
}

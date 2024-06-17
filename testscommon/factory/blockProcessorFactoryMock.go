package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	processBlock "github.com/multiversx/mx-chain-go/process/block"
)

// BlockProcessorFactoryMock -
type BlockProcessorFactoryMock struct {
	CreateBlockProcessorCalled func(argumentsBaseProcessor processBlock.ArgBaseProcessor, argsMetaProcessor processBlock.ArgMetaProcessor) (process.DebuggerBlockProcessor, error)
}

// CreateBlockProcessor -
func (b *BlockProcessorFactoryMock) CreateBlockProcessor(argumentsBaseProcessor processBlock.ArgBaseProcessor, argsMetaProcessor processBlock.ArgMetaProcessor) (process.DebuggerBlockProcessor, error) {
	if b.CreateBlockProcessorCalled != nil {
		return b.CreateBlockProcessorCalled(argumentsBaseProcessor, argsMetaProcessor)
	}
	return nil, nil
}

// IsInterfaceNil -
func (b *BlockProcessorFactoryMock) IsInterfaceNil() bool {
	return b == nil
}

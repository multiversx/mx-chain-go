package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	processBlock "github.com/multiversx/mx-chain-go/process/block"
)

// BlockProcessorFactoryMock -
type BlockProcessorFactoryMock struct {
	CreateBlockProcessorCalled func(argumentsBaseProcessor processBlock.ArgBaseProcessor) (process.DebuggerBlockProcessor, error)
}

// CreateBlockProcessor -
func (b *BlockProcessorFactoryMock) CreateBlockProcessor(args processBlock.ArgBaseProcessor) (process.DebuggerBlockProcessor, error) {
	if b.CreateBlockProcessorCalled != nil {
		return b.CreateBlockProcessorCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil -
func (b *BlockProcessorFactoryMock) IsInterfaceNil() bool {
	return b == nil
}

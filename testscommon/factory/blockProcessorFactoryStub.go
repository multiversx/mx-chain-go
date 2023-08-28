package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	processBlock "github.com/multiversx/mx-chain-go/process/block"
)

// BlockProcessorFactoryStub -
type BlockProcessorFactoryStub struct {
	CreateBlockProcessorCalled func(argumentsBaseProcessor processBlock.ArgBaseProcessor) (process.DebuggerBlockProcessor, error)
}

// NewBlockProcessorFactoryStub -
func NewBlockProcessorFactoryStub() *BlockProcessorFactoryStub {
	return &BlockProcessorFactoryStub{}
}

// CreateBlockProcessor -
func (b *BlockProcessorFactoryStub) CreateBlockProcessor(args processBlock.ArgBaseProcessor) (process.DebuggerBlockProcessor, error) {
	if b.CreateBlockProcessorCalled != nil {
		return b.CreateBlockProcessorCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil -
func (b *BlockProcessorFactoryStub) IsInterfaceNil() bool {
	return false
}

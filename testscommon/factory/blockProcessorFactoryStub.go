package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	processBlock "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// BlockProcessorFactoryStub -
type BlockProcessorFactoryStub struct {
	CreateBlockProcessorCalled func(argumentsBaseProcessor processBlock.ArgBaseProcessor) (process.DebuggerBlockProcessor, error)
}

// CreateBlockProcessor -
func (b *BlockProcessorFactoryStub) CreateBlockProcessor(args processBlock.ArgBaseProcessor) (process.DebuggerBlockProcessor, error) {
	if b.CreateBlockProcessorCalled != nil {
		return b.CreateBlockProcessorCalled(args)
	}
	return &testscommon.BlockProcessorStub{}, nil
}

// IsInterfaceNil -
func (b *BlockProcessorFactoryStub) IsInterfaceNil() bool {
	return b == nil
}

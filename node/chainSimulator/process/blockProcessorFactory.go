package process

import (
	"github.com/multiversx/mx-chain-go/process"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

type blockProcessorFactory struct {
}

// NewBlockProcessorFactory creates a new block processor factory for normal chain simulator
func NewBlockProcessorFactory() BlocksProcessorFactory {
	return &blockProcessorFactory{}
}

// ProcessBlock will create the block in normal chain simulator
func (bpf *blockProcessorFactory) ProcessBlock(blockProcessor process.BlockProcessor, header data.HeaderHandler) (data.HeaderHandler, data.BodyHandler, error) {
	if check.IfNil(blockProcessor) {
		return nil, nil, process.ErrNilBlockProcessor
	}
	if check.IfNil(header) {
		return nil, nil, process.ErrNilHeaderHandler
	}

	return blockProcessor.CreateBlock(header, func() bool {
		return true
	})
}

// IsInterfaceNil returns true if there is no value under the interface
func (bpf *blockProcessorFactory) IsInterfaceNil() bool {
	return bpf == nil
}

package process

import (
	"time"

	chainSimulatorProcess "github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/process"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

type sovereignBlockProcessorFactory struct {
}

// NewSovereignBlockProcessorFactory creates a new block processor factory for sovereign chain simulator
func NewSovereignBlockProcessorFactory() chainSimulatorProcess.BlocksProcessorFactory {
	return &sovereignBlockProcessorFactory{}
}

// ProcessBlock will create and process a block in sovereign chain simulator
func (sbpf *sovereignBlockProcessorFactory) ProcessBlock(blockProcessor process.BlockProcessor, header data.HeaderHandler) (data.HeaderHandler, data.BodyHandler, error) {
	if check.IfNil(blockProcessor) {
		return nil, nil, process.ErrNilBlockProcessor
	}
	if check.IfNil(header) {
		return nil, nil, process.ErrNilHeaderHandler
	}

	header, block, err := blockProcessor.CreateBlock(header, func() bool {
		return true
	})
	if err != nil {
		return nil, nil, err
	}

	return blockProcessor.ProcessBlock(header, block, func() time.Duration {
		return time.Second
	})
}

// IsInterfaceNil returns true if there is no value under the interface
func (sbpf *sovereignBlockProcessorFactory) IsInterfaceNil() bool {
	return sbpf == nil
}

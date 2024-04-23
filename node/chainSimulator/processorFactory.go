package chainSimulator

import (
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
)

type processorFactory struct {
}

// NewProcessorFactory creates a new processor factory for normal chain simulator
func NewProcessorFactory() ChainProcessorFactory {
	return &processorFactory{}
}

// CreateChainHandler creates a new chain handler for normal chain simulator
func (pf *processorFactory) CreateChainHandler(nodeHandler process.NodeHandler) (ChainHandler, error) {
	return process.NewBlocksCreator(nodeHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (pf *processorFactory) IsInterfaceNil() bool {
	return pf == nil
}

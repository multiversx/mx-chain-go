package chainSimulator

import (
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
)

type processorFactory struct {
}

// NewChainHandlerFactory creates a new chain handler factory for normal chain simulator
func NewChainHandlerFactory() ChainHandlerFactory {
	return &processorFactory{}
}

// CreateChainHandler creates a new chain handler for normal chain simulator
func (pf *processorFactory) CreateChainHandler(nodeHandler process.NodeHandler) (ChainHandler, error) {
	return process.NewBlocksCreator(nodeHandler, process.NewBlockProcessorFactory())
}

// IsInterfaceNil returns true if there is no value under the interface
func (pf *processorFactory) IsInterfaceNil() bool {
	return pf == nil
}

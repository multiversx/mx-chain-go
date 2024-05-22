package chainSimulator

import (
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
)

type chainFactory struct {
}

// NewChainFactory creates a new chain factory for normal chain simulator
func NewChainFactory() ChainFactory {
	return &chainFactory{}
}

// CreateChainHandler creates a new chain handler for normal chain simulator
func (pf *chainFactory) CreateChainHandler(nodeHandler process.NodeHandler) (ChainHandler, error) {
	return process.NewBlocksCreator(nodeHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (pf *chainFactory) IsInterfaceNil() bool {
	return pf == nil
}

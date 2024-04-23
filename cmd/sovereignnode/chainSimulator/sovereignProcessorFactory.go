package chainSimulator

import (
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	sovSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator/process"
)

type sovereignProcessorFactory struct {
}

// NewSovereignProcessorFactory creates a new processor factory for normal chain simulator
func NewSovereignProcessorFactory() chainSimulator.ChainProcessorFactory {
	return &sovereignProcessorFactory{}
}

// CreateChainHandler creates a new chain handler for normal chain simulator
func (spf *sovereignProcessorFactory) CreateChainHandler(nodeHandler process.NodeHandler) (chainSimulator.ChainHandler, error) {
	return sovSimulator.NewSovereignBlocksCreator(nodeHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (spf *sovereignProcessorFactory) IsInterfaceNil() bool {
	return spf == nil
}

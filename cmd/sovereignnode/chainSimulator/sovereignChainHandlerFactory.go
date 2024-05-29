package chainSimulator

import (
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	sovProcess "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator/process"
)

type sovereignProcessorFactory struct {
}

// NewSovereignChainHandlerFactory creates a new chain handler factory for sovereign chain simulator
func NewSovereignChainHandlerFactory() chainSimulator.ChainHandlerFactory {
	return &sovereignProcessorFactory{}
}

// CreateChainHandler creates a new chain handler for sovereign chain simulator
func (spf *sovereignProcessorFactory) CreateChainHandler(nodeHandler process.NodeHandler) (chainSimulator.ChainHandler, error) {
	return process.NewBlocksCreator(nodeHandler, sovProcess.NewSovereignBlockProcessorFactory())
}

// IsInterfaceNil returns true if there is no value under the interface
func (spf *sovereignProcessorFactory) IsInterfaceNil() bool {
	return spf == nil
}

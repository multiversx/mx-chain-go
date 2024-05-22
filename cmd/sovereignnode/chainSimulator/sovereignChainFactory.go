package chainSimulator

import (
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	sovSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator/process"
)

type sovereignChainFactory struct {
}

// NewSovereignChainFactory creates a new chain factory for normal chain simulator
func NewSovereignChainFactory() chainSimulator.ChainFactory {
	return &sovereignChainFactory{}
}

// CreateChainHandler creates a new chain handler for normal chain simulator
func (spf *sovereignChainFactory) CreateChainHandler(nodeHandler process.NodeHandler) (chainSimulator.ChainHandler, error) {
	return sovSimulator.NewSovereignBlocksCreator(nodeHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (spf *sovereignChainFactory) IsInterfaceNil() bool {
	return spf == nil
}

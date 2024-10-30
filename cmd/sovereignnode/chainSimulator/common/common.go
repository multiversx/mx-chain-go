package common

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
)

// GetCurrentSovereignHeader returns current sovereign chain block handler from blockchain hook
func GetCurrentSovereignHeader(nodeHandler process.NodeHandler) data.SovereignChainHeaderHandler {
	return nodeHandler.GetChainHandler().GetCurrentBlockHeader().(data.SovereignChainHeaderHandler)
}

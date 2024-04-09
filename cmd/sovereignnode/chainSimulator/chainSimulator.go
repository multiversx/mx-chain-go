package chainSimulator

import (
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
)

// NewSovereignChainSimulator will create a new instance of sovereign chain simulator
func NewSovereignChainSimulator(args chainSimulator.ArgsChainSimulator) (*chainSimulator.Simulator, error) {
	return chainSimulator.NewChainSimulator(args)
}

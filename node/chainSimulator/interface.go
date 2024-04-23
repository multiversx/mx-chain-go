package chainSimulator

import "github.com/multiversx/mx-chain-go/node/chainSimulator/process"

type ChainProcessorFactory interface {
	CreateChainHandler(nodeHandler process.NodeHandler) (ChainHandler, error)
	IsInterfaceNil() bool
}

// ChainHandler defines what a chain handler should be able to do
type ChainHandler interface {
	IncrementRound()
	CreateNewBlock() error
	IsInterfaceNil() bool
}

// ChainSimulator defines what a chain simulator should be able to do
type ChainSimulator interface {
	GenerateBlocks(numOfBlocks int) error
	GetNodeHandler(shardID uint32) process.NodeHandler
	IsInterfaceNil() bool
}

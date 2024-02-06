package helpers

import "github.com/multiversx/mx-chain-go/node/chainSimulator/process"

// ChainSimulator defines what a chain simulator should be able to do
type ChainSimulator interface {
	GenerateBlocks(numOfBlocks int) error
	GetNodeHandler(shardID uint32) process.NodeHandler
	AddValidatorKeys(validatorsPrivateKeys [][]byte) error
	IsInterfaceNil() bool
}

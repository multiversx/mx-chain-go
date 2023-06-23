package hooks

import "github.com/multiversx/mx-chain-go/process"

// blockChainHookFactory - factory for blockchain hook chain run type normal
type blockChainHookFactory struct {
}

// NewBlockChainHookFactory creates a new instance of blockChainHookFactory
func NewBlockChainHookFactory() (BlockChainHookFactoryHandler, error) {
	return &blockChainHookFactory{}, nil
}

// CreateBlockChainHook creates a blockchain hook based on the chain run type normal
func (bhf *blockChainHookFactory) CreateBlockChainHook(args ArgBlockChainHook) (process.BlockChainHookHandler, error) {
	return NewBlockChainHookImpl(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (bhf *blockChainHookFactory) IsInterfaceNil() bool {
	return bhf == nil
}

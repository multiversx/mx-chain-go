package hooks

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	customErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
)

// sovereignBlockChainHookFactory - factory for blockchain hook chain run type sovereign
type sovereignBlockChainHookFactory struct {
	blockChainHookFactory BlockChainHookFactoryHandler
}

// NewSovereignBlockChainHookFactory creates a new instance of sovereignBlockChainHookFactory
func NewSovereignBlockChainHookFactory(blockChainHookFactory BlockChainHookFactoryHandler) (BlockChainHookFactoryHandler, error) {
	if check.IfNil(blockChainHookFactory) {
		return nil, customErrors.ErrNilBlockChainHookFactory
	}
	return &sovereignBlockChainHookFactory{
		blockChainHookFactory: blockChainHookFactory,
	}, nil
}

// CreateBlockChainHook creates a blockchain hook based on the chain run type sovereign
func (bhf *sovereignBlockChainHookFactory) CreateBlockChainHook(args ArgBlockChainHook) (process.BlockChainHookHandler, error) {
	bh, _ := NewBlockChainHookImpl(args)
	return NewSovereignBlockChainHook(bh)
}

// IsInterfaceNil returns true if there is no value under the interface
func (bhf *sovereignBlockChainHookFactory) IsInterfaceNil() bool {
	return bhf == nil
}

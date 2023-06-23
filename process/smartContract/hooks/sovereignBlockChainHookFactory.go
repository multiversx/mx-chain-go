package hooks

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	customErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
)

// SovereignBlockChainHookFactory - factory for blockchain hook chain run type sovereign
type SovereignBlockChainHookFactory struct {
	blockChainHookFactory BlockChainHookFactoryHandler
}

// NewSovereignBlockChainHookFactory creates a new instance of SovereignBlockChainHookFactory
func NewSovereignBlockChainHookFactory(blockChainHookFactory BlockChainHookFactoryHandler) (BlockChainHookFactoryHandler, error) {
	if check.IfNil(blockChainHookFactory) {
		return nil, customErrors.ErrNilBlockChainHookFactory
	}
	return &SovereignBlockChainHookFactory{
		blockChainHookFactory: blockChainHookFactory,
	}, nil
}

// CreateBlockChainHook creates a blockchain hook based on the chain run type sovereign
func (bhf *SovereignBlockChainHookFactory) CreateBlockChainHook(args ArgBlockChainHook) (process.BlockChainHookHandler, error) {
	bh, _ := NewBlockChainHookImpl(args)
	return NewSovereignBlockChainHook(bh)
}

// IsInterfaceNil returns true if there is no value under the interface
func (bhf *SovereignBlockChainHookFactory) IsInterfaceNil() bool {
	return bhf == nil
}

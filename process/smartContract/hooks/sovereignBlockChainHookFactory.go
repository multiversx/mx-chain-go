package hooks

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	customErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
)

// sovereignBlockChainHookFactory - factory for blockchain hook chain run type sovereign
type sovereignBlockChainHookFactory struct {
	blockChainHookFactory BlockChainHookHandlerCreator
}

// NewSovereignBlockChainHookFactory creates a new instance of sovereignBlockChainHookFactory
func NewSovereignBlockChainHookFactory(blockChainHookFactory BlockChainHookHandlerCreator) (BlockChainHookHandlerCreator, error) {
	if check.IfNil(blockChainHookFactory) {
		return nil, customErrors.ErrNilBlockChainHookFactory
	}
	return &sovereignBlockChainHookFactory{
		blockChainHookFactory: blockChainHookFactory,
	}, nil
}

// CreateBlockChainHookHandler creates a blockchain hook based on the chain run type sovereign
func (bhf *sovereignBlockChainHookFactory) CreateBlockChainHookHandler(args ArgBlockChainHook) (process.BlockChainHookWithAccountsAdapter, error) {
	bh, err := NewBlockChainHookImpl(args)
	if err != nil {
		return nil, err
	}
	return NewSovereignBlockChainHook(bh)
}

// IsInterfaceNil returns true if there is no value under the interface
func (bhf *sovereignBlockChainHookFactory) IsInterfaceNil() bool {
	return bhf == nil
}

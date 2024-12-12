package hooks

import (
	"github.com/multiversx/mx-chain-go/process"
)

// sovereignBlockChainHookFactory - factory for blockchain hook chain run type sovereign
type sovereignBlockChainHookFactory struct {
}

// NewSovereignBlockChainHookFactory creates a new instance of sovereignBlockChainHookFactory
func NewSovereignBlockChainHookFactory() BlockChainHookHandlerCreator {
	return &sovereignBlockChainHookFactory{}
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

package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
)

// BlockChainHookHandlerFactoryStub -
type BlockChainHookHandlerFactoryStub struct {
	CreateBlockChainHookHandlerCalled func(args hooks.ArgBlockChainHook) (process.BlockChainHookHandler, error)
}

// CreateBlockChainHookHandler -
func (b *BlockChainHookHandlerFactoryStub) CreateBlockChainHookHandler(args hooks.ArgBlockChainHook) (process.BlockChainHookHandler, error) {
	if b.CreateBlockChainHookHandlerCalled != nil {
		return b.CreateBlockChainHookHandlerCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil -
func (b *BlockChainHookHandlerFactoryStub) IsInterfaceNil() bool {
	return b == nil
}

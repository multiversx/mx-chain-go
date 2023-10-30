package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// BlockChainHookHandlerFactoryMock -
type BlockChainHookHandlerFactoryMock struct {
	CreateBlockChainHookHandlerCalled func(args hooks.ArgBlockChainHook) (process.BlockChainHookHandler, error)
}

// CreateBlockChainHookHandler -
func (b *BlockChainHookHandlerFactoryMock) CreateBlockChainHookHandler(args hooks.ArgBlockChainHook) (process.BlockChainHookHandler, error) {
	if b.CreateBlockChainHookHandlerCalled != nil {
		return b.CreateBlockChainHookHandlerCalled(args)
	}
	return &testscommon.BlockChainHookStub{}, nil
}

// IsInterfaceNil -
func (b *BlockChainHookHandlerFactoryMock) IsInterfaceNil() bool {
	return b == nil
}

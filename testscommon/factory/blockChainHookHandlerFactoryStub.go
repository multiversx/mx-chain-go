package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// BlockChainHookHandlerFactorySub -
type BlockChainHookHandlerFactorySub struct {
}

// CreateBlockChainHookHandler -
func (b *BlockChainHookHandlerFactorySub) CreateBlockChainHookHandler(args hooks.ArgBlockChainHook) (process.BlockChainHookHandler, error) {
	return &testscommon.BlockChainHookStub{}, nil
}

// IsInterfaceNil -
func (b *BlockChainHookHandlerFactorySub) IsInterfaceNil() bool {
	return b == nil
}

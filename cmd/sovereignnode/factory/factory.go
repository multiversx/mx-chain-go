package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
)

// SovereignBlockChainHookFactory - factory for sovereign run
type SovereignBlockChainHookFactory struct {
}

func (bhf *SovereignBlockChainHookFactory) CreateBlockChainHook(args hooks.ArgBlockChainHook) (process.BlockChainHookHandler, error) {
	bh, _ := hooks.NewBlockChainHookImpl(args)
	return hooks.NewSovereignBlockChainHook(bh)
}

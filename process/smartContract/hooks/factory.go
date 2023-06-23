package hooks

import (
	"fmt"
	"github.com/multiversx/mx-chain-go/common"
	customErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
)

// BlockChainHookFactoryHandler defines the blockchain hook factory handler
type BlockChainHookFactoryHandler interface {
	CreateBlockChainHook(args ArgBlockChainHook) (process.BlockChainHookHandler, error)
}

// blockChainHookFactory - factory for blockchain hook chain run type normal
type blockChainHookFactory struct {
}

// NewBlockChainHookFactory creates a new instance of blockChainHookFactory
func NewBlockChainHookFactory() BlockChainHookFactoryHandler {
	return &blockChainHookFactory{}
}

// CreateBlockChainHook creates a blockchain hook based on the chain run type normal
func (bhf *blockChainHookFactory) CreateBlockChainHook(args ArgBlockChainHook) (process.BlockChainHookHandler, error) {
	return NewBlockChainHookImpl(args)
}

// SovereignBlockChainHookFactory - factory for blockchain hook chain run type sovereign
type SovereignBlockChainHookFactory struct {
	blockChainHookFactory BlockChainHookFactoryHandler
}

// NewSovereignBlockChainHookFactory creates a new instance of SovereignBlockChainHookFactory
func NewSovereignBlockChainHookFactory(blockChainHookFactory BlockChainHookFactoryHandler) BlockChainHookFactoryHandler {
	return &SovereignBlockChainHookFactory{
		blockChainHookFactory: blockChainHookFactory,
	}
}

// CreateBlockChainHook creates a blockchain hook based on the chain run type sovereign
func (bhf *SovereignBlockChainHookFactory) CreateBlockChainHook(args ArgBlockChainHook) (process.BlockChainHookHandler, error) {
	bh, _ := NewBlockChainHookImpl(args)
	return NewSovereignBlockChainHook(bh)
}

// CreateBlockChainHook creates a blockchain hook based on the chain run type (normal/sovereign)
func CreateBlockChainHook(chainRunType common.ChainRunType, args ArgBlockChainHook) (process.BlockChainHookHandler, error) {
	switch chainRunType {
	case common.ChainRunTypeRegular:
		return NewBlockChainHookFactory().CreateBlockChainHook(args)
	case common.ChainRunTypeSovereign:
		return NewSovereignBlockChainHookFactory(NewBlockChainHookFactory()).CreateBlockChainHook(args)
	default:
		return nil, fmt.Errorf("%w type %v", customErrors.ErrUnimplementedChainRunType, chainRunType)
	}
}

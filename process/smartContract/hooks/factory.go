package hooks

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/common"
	customErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
)

type BlockChainHookFactoryHandler interface {
	CreateBlockChainHook(args ArgBlockChainHook) (process.BlockChainHookHandler, error)
}

// BlockChainHookFactory - factory for normal run
type BlockChainHookFactory struct {
}

func (bhf *BlockChainHookFactory) CreateBlockChainHook(args ArgBlockChainHook) (process.BlockChainHookHandler, error) {
	return NewBlockChainHookImpl(args)
}

// SovereignBlockChainHookFactory - factory for sovereign run
type SovereignBlockChainHookFactory struct {
}

func (bhf *SovereignBlockChainHookFactory) CreateBlockChainHook(args ArgBlockChainHook) (process.BlockChainHookHandler, error) {
	bh, _ := NewBlockChainHookImpl(args)
	return NewSovereignBlockChainHook(bh)
}

// CreateBlockChainHook creates a blockchain hook based on the chain run type (normal/sovereign)
func CreateBlockChainHook(chainRunType common.ChainRunType, args ArgBlockChainHook) (process.BlockChainHookHandler, error) {
	bh, err := NewBlockChainHookImpl(args)

	switch chainRunType {
	case common.ChainRunTypeRegular:
		return bh, err
	case common.ChainRunTypeSovereign:
		return NewSovereignBlockChainHook(bh)
	default:
		return nil, fmt.Errorf("%w type %v", customErrors.ErrUnimplementedChainRunType, chainRunType)
	}
}

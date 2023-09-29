package hooks

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/common"
	customErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
)

// CreateBlockChainHook creates a blockchain hook based on the chain run type (normal/sovereign)
func CreateBlockChainHook(chainRunType common.ChainRunType, args ArgBlockChainHook) (process.BlockChainHookWithAccountsAdapter, error) {
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

package hooks

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/common"
	customErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
)

// CreateBlockChainHook creates a blockchain hook based on the chain run type (normal/sovereign)
func CreateBlockChainHook(chainRunType common.ChainRunType, args ArgBlockChainHook) (process.BlockChainHookHandler, error) {
	factory, err := NewBlockChainHookFactory()
	if err != nil {
		return nil, err
	}

	var bhhc BlockChainHookHandlerCreator

	switch chainRunType {
	case common.ChainRunTypeRegular:
		bhhc = factory
	case common.ChainRunTypeSovereign:
		sovereignFactory, sovErr := NewSovereignBlockChainHookFactory(factory)
		if sovErr != nil {
			return nil, sovErr
		}
		bhhc = sovereignFactory
	default:
		return nil, fmt.Errorf("%w type %v", customErrors.ErrUnimplementedChainRunType, chainRunType)
	}

	return bhhc.CreateBlockChainHookHandler(args)
}

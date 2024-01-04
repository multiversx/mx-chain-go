package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
	"github.com/multiversx/mx-chain-go/node/trieIterators/disabled"
)

type delegatedListHandlerFactory struct {
}

func NewDelegatedListHandlerFactory() *delegatedListHandlerFactory {
	return &delegatedListHandlerFactory{}
}

// CreateDelegatedListHandler will create a new instance of DirectStakedListHandler
func (d *delegatedListHandlerFactory) CreateDelegatedListHandler(args trieIterators.ArgTrieIteratorProcessor) (external.DelegatedListHandler, error) {
	//TODO add unit tests
	if args.ShardID != core.MetachainShardId {
		return disabled.NewDisabledDelegatedListProcessor(), nil
	}

	return trieIterators.NewDelegatedListProcessor(args)
}

func (d *delegatedListHandlerFactory) IsInterfaceNil() bool {
	return d == nil
}

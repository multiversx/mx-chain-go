package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
	"github.com/multiversx/mx-chain-go/node/trieIterators/disabled"
)

type sovereignDirectStakedListHandlerFactory struct {
}

func NewSovereignDirectStakedListHandlerFactory() *sovereignDirectStakedListHandlerFactory {
	return &sovereignDirectStakedListHandlerFactory{}
}

// CreateDirectStakedListHandler will create a new instance of DirectStakedListHandler
func (ds *sovereignDirectStakedListHandlerFactory) CreateDirectStakedListHandler(args trieIterators.ArgTrieIteratorProcessor) (external.DirectStakedListHandler, error) {
	//TODO add unit tests
	if args.ShardID != core.MetachainShardId {
		return disabled.NewDisabledDirectStakedListProcessor(), nil
	}

	return trieIterators.NewDirectStakedListProcessor(args)
}

func (ds *sovereignDirectStakedListHandlerFactory) IsInterfaceNil() bool {
	return ds == nil
}

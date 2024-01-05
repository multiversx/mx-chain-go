package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
	"github.com/multiversx/mx-chain-go/node/trieIterators/disabled"
)

type directStakedListHandlerFactory struct {
}

// NewDirectStakedListHandlerFactory create a new direct staked list handler
func NewDirectStakedListHandlerFactory() *directStakedListHandlerFactory {
	return &directStakedListHandlerFactory{}
}

// CreateDirectStakedListHandler will create a new instance of DirectStakedListHandler
func (ds *directStakedListHandlerFactory) CreateDirectStakedListHandler(args trieIterators.ArgTrieIteratorProcessor) (external.DirectStakedListHandler, error) {
	if args.ShardID != core.MetachainShardId {
		return disabled.NewDisabledDirectStakedListProcessor(), nil
	}

	return trieIterators.NewDirectStakedListProcessor(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ds *directStakedListHandlerFactory) IsInterfaceNil() bool {
	return ds == nil
}

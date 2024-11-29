package factory

import (
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
)

// DelegatedListProcessorFactoryHandler can create delegated list processor handler
type DelegatedListProcessorFactoryHandler interface {
	CreateDelegatedListProcessorHandler(args trieIterators.ArgTrieIteratorProcessor) (external.DelegatedListHandler, error)
	IsInterfaceNil() bool
}

// DirectStakedListProcessorFactoryHandler can create direct staked list processor handler
type DirectStakedListProcessorFactoryHandler interface {
	CreateDirectStakedListProcessorHandler(args trieIterators.ArgTrieIteratorProcessor) (external.DirectStakedListHandler, error)
	IsInterfaceNil() bool
}

// TotalStakedValueProcessorFactoryHandler can create total staked value processor handler
type TotalStakedValueProcessorFactoryHandler interface {
	CreateTotalStakedValueProcessorHandler(args trieIterators.ArgTrieIteratorProcessor) (external.TotalStakedValueHandler, error)
	IsInterfaceNil() bool
}

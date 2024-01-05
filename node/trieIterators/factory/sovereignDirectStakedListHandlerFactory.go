package factory

import (
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
)

type sovereignDirectStakedListHandlerFactory struct {
}

// NewSovereignDirectStakedListHandlerFactory create a new sovereign direct staked list handler
func NewSovereignDirectStakedListHandlerFactory() *sovereignDirectStakedListHandlerFactory {
	return &sovereignDirectStakedListHandlerFactory{}
}

// CreateDirectStakedListHandler will create a new instance of DirectStakedListHandler
func (ds *sovereignDirectStakedListHandlerFactory) CreateDirectStakedListHandler(args trieIterators.ArgTrieIteratorProcessor) (external.DirectStakedListHandler, error) {
	return trieIterators.NewDirectStakedListProcessor(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ds *sovereignDirectStakedListHandlerFactory) IsInterfaceNil() bool {
	return ds == nil
}

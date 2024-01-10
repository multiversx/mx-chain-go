package factory

import (
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
)

type sovereignDirectStakedListHandlerFactory struct {
}

// NewSovereignDirectStakedListProcessorFactory creates a new sovereign direct staked list handler
func NewSovereignDirectStakedListProcessorFactory() *sovereignDirectStakedListHandlerFactory {
	return &sovereignDirectStakedListHandlerFactory{}
}

// CreateDirectStakedListProcessorHandler creates a new instance of DirectStakedListHandler
func (ds *sovereignDirectStakedListHandlerFactory) CreateDirectStakedListProcessorHandler(args trieIterators.ArgTrieIteratorProcessor) (external.DirectStakedListHandler, error) {
	return trieIterators.NewDirectStakedListProcessor(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ds *sovereignDirectStakedListHandlerFactory) IsInterfaceNil() bool {
	return ds == nil
}

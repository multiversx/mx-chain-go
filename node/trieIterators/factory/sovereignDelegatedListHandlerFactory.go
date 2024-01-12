package factory

import (
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
)

type sovereignDelegatedListProcessorFactory struct {
}

// NewSovereignDelegatedListProcessorFactory creates a new sovereign delegated list handler
func NewSovereignDelegatedListProcessorFactory() *sovereignDelegatedListProcessorFactory {
	return &sovereignDelegatedListProcessorFactory{}
}

// CreateDelegatedListProcessorHandler creates a new instance of delegated list processor for sovereign chains
func (sd *sovereignDelegatedListProcessorFactory) CreateDelegatedListProcessorHandler(args trieIterators.ArgTrieIteratorProcessor) (external.DelegatedListHandler, error) {
	return trieIterators.NewDelegatedListProcessor(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (sd *sovereignDelegatedListProcessorFactory) IsInterfaceNil() bool {
	return sd == nil
}

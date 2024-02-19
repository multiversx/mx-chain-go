package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
	"github.com/multiversx/mx-chain-go/node/trieIterators/disabled"
)

type delegatedListProcessorFactory struct {
}

// NewDelegatedListProcessorFactory create a new delegated list handler
func NewDelegatedListProcessorFactory() *delegatedListProcessorFactory {
	return &delegatedListProcessorFactory{}
}

// CreateDelegatedListProcessorHandler will create a new instance of delegated list processor for regular/normal chain
func (d *delegatedListProcessorFactory) CreateDelegatedListProcessorHandler(args trieIterators.ArgTrieIteratorProcessor) (external.DelegatedListHandler, error) {
	if args.ShardID != core.MetachainShardId {
		return disabled.NewDisabledDelegatedListProcessor(), nil
	}

	return trieIterators.NewDelegatedListProcessor(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (d *delegatedListProcessorFactory) IsInterfaceNil() bool {
	return d == nil
}

package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
	"github.com/multiversx/mx-chain-go/node/trieIterators/disabled"
)

type directStakedListProcessorFactory struct {
}

// NewDirectStakedListProcessorFactory create a new direct staked list handler
func NewDirectStakedListProcessorFactory() *directStakedListProcessorFactory {
	return &directStakedListProcessorFactory{}
}

// CreateDirectStakedListProcessorHandler will create a new instance of direct staked list processor for regular/normal chain
func (ds *directStakedListProcessorFactory) CreateDirectStakedListProcessorHandler(args trieIterators.ArgTrieIteratorProcessor) (external.DirectStakedListHandler, error) {
	if args.ShardID != core.MetachainShardId {
		return disabled.NewDisabledDirectStakedListProcessor(), nil
	}

	return trieIterators.NewDirectStakedListProcessor(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ds *directStakedListProcessorFactory) IsInterfaceNil() bool {
	return ds == nil
}

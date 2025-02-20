package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
	"github.com/multiversx/mx-chain-go/node/trieIterators/disabled"
)

type totalStakedValueProcessorFactory struct {
}

// NewTotalStakedListProcessorFactory creates a new total staked value handler
func NewTotalStakedListProcessorFactory() *totalStakedValueProcessorFactory {
	return &totalStakedValueProcessorFactory{}
}

// CreateTotalStakedValueProcessorHandler will create a new instance of total staked value processor for regular/normal chain
func (d *totalStakedValueProcessorFactory) CreateTotalStakedValueProcessorHandler(args trieIterators.ArgTrieIteratorProcessor) (external.TotalStakedValueHandler, error) {
	if args.ShardID != core.MetachainShardId {
		return disabled.NewDisabledStakeValuesProcessor(), nil
	}

	return trieIterators.NewTotalStakedValueProcessor(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (d *totalStakedValueProcessorFactory) IsInterfaceNil() bool {
	return d == nil
}

package factory

import (
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
)

// CreateTotalStakedValueHandler will create a new instance of TotalStakedValueHandler
func CreateTotalStakedValueHandler(args trieIterators.ArgTrieIteratorProcessor) (external.TotalStakedValueHandler, error) {
	// TODO: based on chain type, allow shards nodes as well
	//if args.ShardID != core.MetachainShardId {
	//	return disabled.NewDisabledStakeValuesProcessor(), nil
	//}

	return trieIterators.NewTotalStakedValueProcessor(args)
}

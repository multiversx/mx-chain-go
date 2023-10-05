package factory

import (
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
)

// CreateDelegatedListHandler will create a new instance of DirectStakedListHandler
func CreateDelegatedListHandler(args trieIterators.ArgTrieIteratorProcessor) (external.DelegatedListHandler, error) {
	//TODO add unit tests
	//TODO2: only if the chain type is sovereign, then allow shard nodes no respond to this endpoint
	//if args.ShardID != core.MetachainShardId {
	//	return disabled.NewDisabledDelegatedListProcessor(), nil
	//}

	return trieIterators.NewDelegatedListProcessor(args)
}

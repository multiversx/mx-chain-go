package factory

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/trieIterators"
	"github.com/ElrondNetwork/elrond-go/node/trieIterators/disabled"
)

// CreateDelegatedListHandler will create a new instance of DirectStakedListHandler
func CreateDelegatedListHandler(args trieIterators.ArgTrieIteratorProcessor) (external.DelegatedListHandler, error) {
	//TODO add unit tests
	if args.ShardID != core.MetachainShardId {
		return disabled.NewDisabledDelegatedListProcessor(), nil
	}

	return trieIterators.NewDelegatedListProcessor(args)
}

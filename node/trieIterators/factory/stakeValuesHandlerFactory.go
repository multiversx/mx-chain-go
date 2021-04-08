package factory

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/trieIterators"
	"github.com/ElrondNetwork/elrond-go/node/trieIterators/disabled"
)

// CreateTotalStakedValueHandler will create a new instance of TotalStakedValueHandler
func CreateTotalStakedValueHandler(args trieIterators.ArgTrieIteratorProcessor) (external.TotalStakedValueHandler, error) {
	if args.ShardID != core.MetachainShardId {
		return disabled.NewDisabledStakeValuesProcessor(), nil
	}

	return trieIterators.NewTotalStakedValueProcessor(args)
}

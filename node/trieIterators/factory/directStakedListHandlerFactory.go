package factory

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/trieIterators"
	"github.com/ElrondNetwork/elrond-go/node/trieIterators/disabled"
)

// CreateDirectStakedListHandler will create a new instance of DirectStakedListHandler
func CreateDirectStakedListHandler(args *ArgStakeProcessors) (external.DirectStakedListHandler, error) {
	//TODO add unit tests
	if args.ShardID != core.MetachainShardId {
		return disabled.NewDisabledDirectStakedListProcessor(), nil
	}

	return trieIterators.NewDirectStakedListProcessor(
		args.Accounts,
		args.BlockChain,
		args.QueryService,
		args.PublicKeyConverter,
	)
}

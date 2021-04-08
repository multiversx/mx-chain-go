package factory

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/trieIterators"
	"github.com/ElrondNetwork/elrond-go/node/trieIterators/disabled"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ArgStakeProcessors is a struct that contains components that are needed to create a TotalStakedValueHandler
type ArgStakeProcessors struct {
	ShardID            uint32
	Accounts           *trieIterators.AccountsWrapper
	BlockChain         data.ChainHandler
	QueryService       process.SCQueryService
	PublicKeyConverter core.PubkeyConverter
}

// CreateTotalStakedValueHandler will create a new instance of TotalStakedValueHandler
func CreateTotalStakedValueHandler(args *ArgStakeProcessors) (external.TotalStakedValueHandler, error) {
	if args.ShardID != core.MetachainShardId {
		return disabled.NewDisabledStakeValuesProcessor(), nil
	}

	return trieIterators.NewTotalStakedValueProcessor(
		args.Accounts,
		args.BlockChain,
		args.QueryService,
		args.PublicKeyConverter,
	)
}

package stakeValuesProcessor

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/stakeValuesProcessor/disabled"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ArgsTotalStakedValueHandler is struct that contains components that are needed to create a TotalStakedValueHandler
type ArgsTotalStakedValueHandler struct {
	ShardID            uint32
	Accounts           state.AccountsAdapter
	BlockChain         data.ChainHandler
	QueryService       process.SCQueryService
	PublicKeyConverter core.PubkeyConverter
}

// CreateTotalStakedValueHandler wil create a new instance of TotalStakedValueHandler
func CreateTotalStakedValueHandler(args *ArgsTotalStakedValueHandler) (external.TotalStakedValueHandler, error) {
	if args.ShardID != core.MetachainShardId {
		return disabled.NewDisabledStakeValuesProcessor()
	}

	return NewTotalStakedValueProcessor(
		args.Accounts,
		args.BlockChain,
		args.QueryService,
		args.PublicKeyConverter,
	)
}

package stakeValuesProcessor

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node/external"
)

// ArgsTotalStakedValueHandler is struct that contains components that are needed to create a TotalStakedValueHandler
type ArgsTotalStakedValueHandler struct {
	ShardID             uint32
	NodePrice           string
	InternalMarshalizer marshal.Marshalizer
	Accounts            state.AccountsAdapter
}

// CreateTotalStakedValueHandler wil create a new instance of TotalStakedValueHandler
func CreateTotalStakedValueHandler(args *ArgsTotalStakedValueHandler) (external.TotalStakedValueHandler, error) {
	if args.ShardID != core.MetachainShardId {
		return NewDisabledStakeValuesProcessor()
	}

	return NewTotalStakedValueProcessor(
		args.NodePrice,
		args.InternalMarshalizer,
		args.Accounts,
	)
}

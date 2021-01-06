package totalStakedAPI

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// ArgsTotalStakedValueHandler is struct that contains components that are needed to create a TotalStakedValueHandler
type ArgsTotalStakedValueHandler struct {
	ShardID                                uint32
	TotalStakedValueCacheDurationInMinutes int
	InternalMarshalizer                    marshal.Marshalizer
	Accounts                               state.AccountsAdapter
}

// CreateTotalStakedValueHandler wil create a new instance of TotalStakedValueHandler
func CreateTotalStakedValueHandler(args *ArgsTotalStakedValueHandler) (TotalStakedValueHandler, error) {
	if args.ShardID != core.MetachainShardId {
		return NewDisabledTotalStakedValueProcessor()
	}

	return NewTotalStakedValueProcessor(
		args.InternalMarshalizer,
		time.Duration(args.TotalStakedValueCacheDurationInMinutes)*time.Minute,
		args.Accounts,
	)
}

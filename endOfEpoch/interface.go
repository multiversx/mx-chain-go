package endOfEpoch

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"time"
)

// TriggerHandler is an interface to notify end of epoch
type TriggerHandler interface {
	ForceEndOfEpoch()
	IsEndOfEpoch() bool
	Epoch() uint32
	ReceivedHeader(header data.HeaderHandler)
	IsInterfaceNil() bool
}

// PendingMiniBlocksHandler is an interface to keep unfinalized miniblocks
type PendingMiniBlocksHandler interface {
	PendingMiniBlockHeaders() []block.ShardMiniBlockHeader
	AddCommittedHeader(handler data.HeaderHandler) error
	RevertHeader(handler data.HeaderHandler) error
	IsInterfaceNil() bool
}

// Rounder defines the actions which should be handled by a round implementation
type Rounder interface {
	Index() int64
	// UpdateRound updates the index and the time stamp of the round depending of the epoch genesis time and the current time given
	UpdateRound(time.Time, time.Time)
	// TimeStamp returns the time stamp of the round
	TimeStamp() time.Time
	IsInterfaceNil() bool
}

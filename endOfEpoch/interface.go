package endOfEpoch

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"time"
)

type TriggerHandler interface {
	ForceEndOfEpoch()
	IsEndOfEpoch() bool
	Epoch() uint32
	ReceivedHeader(header data.HeaderHandler)
	IsInterfaceNil() bool
}

type PendingMiniBlocksHandler interface {
	PendingMiniBlocks() []block.MiniBlockHeader
	AddMiniBlockHeader([]block.MiniBlockHeader)
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

type SettingsHandler interface {
	MinRoundsBetweenEpochs() int64
	RoundsPerEpoch() int64
	IsInterfaceNil() bool
}

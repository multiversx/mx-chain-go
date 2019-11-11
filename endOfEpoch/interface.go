package endOfEpoch

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// TriggerHandler defines the functionalities for an end of epoch trigger
type TriggerHandler interface {
	ForceEndOfEpoch(round int64) error
	IsEndOfEpoch() bool
	Epoch() uint32
	ReceivedHeader(header data.HeaderHandler)
	Update(round int64)
	Processed()
	IsInterfaceNil() bool
}

// PendingMiniBlocksHandler defines the actions which should be handled by pending miniblocks implementation
type PendingMiniBlocksHandler interface {
	PendingMiniBlocks() []block.MiniBlockHeader
	AddMiniBlockHeader([]block.MiniBlockHeader)
	IsInterfaceNil() bool
}

// Rounder defines the actions which should be handled by a round implementation
type Rounder interface {
	// Index return the current round
	Index() int64
	// TimeStamp returns the time stamp of the round
	TimeStamp() time.Time
	IsInterfaceNil() bool
}

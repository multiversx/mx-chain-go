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
	SetIsEndOfEpoch(isEndOfEpoch bool)
	EpochStartRound() uint64
	IsInterfaceNil() bool
}

// PendingMiniBlocksHandler defines the actions which should be handled by pending miniblocks implementation
type PendingMiniBlocksHandler interface {
	PendingMiniBlockHeaders() []block.ShardMiniBlockHeader
	AddProcessedHeader(handler data.HeaderHandler) error
	RevertHeader(handler data.HeaderHandler) error
	IsInterfaceNil() bool
}

// Rounder defines the actions which should be handled by a round implementation
type Rounder interface {
	// Index returns the current round
	Index() int64
	// TimeStamp returns the time stamp of the round
	TimeStamp() time.Time
	IsInterfaceNil() bool
}

// HeaderValidator defines the actions needed to validate a header
type HeaderValidator interface {
	IsHeaderConstructionValid(currHdr, prevHdr data.HeaderHandler) error
	IsInterfaceNil() bool
}

// RequestHandler defines the methods through which request to data can be made
type RequestHandler interface {
	RequestHeaderByNonce(shardId uint32, nonce uint64)
	RequestHeader(shardId uint32, hash []byte)
	IsInterfaceNil() bool
}

package epochStart

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// TriggerHandler defines the functionalities for an start of epoch trigger
type TriggerHandler interface {
	ForceEpochStart(round uint64) error
	IsEpochStart() bool
	Epoch() uint32
	ReceivedHeader(header data.HeaderHandler)
	Update(round uint64)
	EpochStartRound() uint64
	EpochStartMetaHdrHash() []byte
	GetSavedStateKey() []byte
	LoadState(key []byte) error
	SetProcessed(header data.HeaderHandler)
	SetFinalityAttestingRound(round uint64)
	EpochFinalityAttestingRound() uint64
	Revert(round uint64)
	SetCurrentEpochStartRound(round uint64)
	IsInterfaceNil() bool
}

// PendingMiniBlocksHandler defines the actions which should be handled by pending miniblocks implementation
type PendingMiniBlocksHandler interface {
	PendingMiniBlockHeaders(lastNotarizedHeaders []data.HeaderHandler) ([]block.ShardMiniBlockHeader, error)
	AddProcessedHeader(handler data.HeaderHandler) error
	RevertHeader(handler data.HeaderHandler) error
	GetNumPendingMiniBlocksForShard(shardID uint32) uint32
	SetNumPendingMiniBlocksForShard(shardID uint32, numPendingMiniBlocks uint32)
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
	RequestShardHeader(shardId uint32, hash []byte)
	RequestMetaHeader(hash []byte)
	RequestMetaHeaderByNonce(nonce uint64)
	RequestShardHeaderByNonce(shardId uint32, nonce uint64)
	IsInterfaceNil() bool
}

// EpochStartHandler defines the action taken on epoch start event
type EpochStartHandler interface {
	EpochStartAction(hdr data.HeaderHandler)
	EpochStartPrepare(hdr data.HeaderHandler)
}

// EpochStartSubscriber provides Register and Unregister functionality for the end of epoch events
type EpochStartSubscriber interface {
	RegisterHandler(handler EpochStartHandler)
	UnregisterHandler(handler EpochStartHandler)
}

// EpochStartNotifier defines which actions should be done for handling new epoch's events
type EpochStartNotifier interface {
	NotifyAll(hdr data.HeaderHandler)
	NotifyAllPrepare(hdr data.HeaderHandler)
	IsInterfaceNil() bool
}

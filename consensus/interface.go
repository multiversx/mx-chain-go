package consensus

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
)

// Rounder defines the actions which should be handled by a round implementation
type Rounder interface {
	Index() int64
	// UpdateRound updates the index and the time stamp of the round depending of the genesis time and the current time given
	UpdateRound(time.Time, time.Time)
	TimeStamp() time.Time
	TimeDuration() time.Duration
	RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration
	IsInterfaceNil() bool
}

// SubroundHandler defines the actions which should be handled by a subround implementation
type SubroundHandler interface {
	// DoWork implements of the subround's job
	DoWork(rounder Rounder) bool
	// Previous returns the ID of the previous subround
	Previous() int
	// Next returns the ID of the next subround
	Next() int
	// Current returns the ID of the current subround
	Current() int
	// StartTime returns the start time, in the rounder time, of the current subround
	StartTime() int64
	// EndTime returns the top limit time, in the rounder time, of the current subround
	EndTime() int64
	// Name returns the name of the current rounder
	Name() string
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// ChronologyHandler defines the actions which should be handled by a chronology implementation
type ChronologyHandler interface {
	AddSubround(SubroundHandler)
	RemoveAllSubrounds()
	// StartRounds starts rounds in a sequential manner, one after the other
	StartRounds()
	Epoch() uint32
	IsInterfaceNil() bool
}

// EpochStartHandler defines that actions which should be handled by an start of epoch implementation
type EpochStartHandler interface {
	IsEpochStart() bool
	Update(round int64)
	Epoch() uint32
	IsInterfaceNil() bool
}

// SposFactory defines an interface for a consensus implementation
type SposFactory interface {
	GenerateSubrounds()
	IsInterfaceNil() bool
}

// BroadcastMessenger defines the behaviour of the broadcast messages by the consensus group
type BroadcastMessenger interface {
	BroadcastBlock(data.BodyHandler, data.HeaderHandler) error
	BroadcastHeader(data.HeaderHandler) error
	BroadcastShardHeader(data.HeaderHandler) error
	BroadcastMiniBlocks(map[uint32][]byte) error
	BroadcastTransactions(map[string][][]byte) error
	BroadcastConsensusMessage(*Message) error
	IsInterfaceNil() bool
}

// P2PMessenger defines a subset of the p2p.Messenger interface
type P2PMessenger interface {
	Broadcast(topic string, buff []byte)
	IsInterfaceNil() bool
}

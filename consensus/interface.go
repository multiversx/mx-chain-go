package consensus

import (
	"math/big"
	"time"
)

// Rounder defines the actions which should be handled by a round implementation
type Rounder interface {
	Index() int32
	// UpdateRound updates the index and the time stamp of the round depending of the genesis time and the current time given
	UpdateRound(time.Time, time.Time)
	TimeStamp() time.Time
	TimeDuration() time.Duration
	RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration
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
}

// ChronologyHandler defines the actions which should be handled by a chronology implementation
type ChronologyHandler interface {
	AddSubround(SubroundHandler)
	RemoveAllSubrounds()
	// StartRounds starts rounds in a sequential manner, one after the other
	StartRounds()
}

// SposFactory defines an interface for a consensus implementation
type SposFactory interface {
	GenerateSubrounds()
}

// Validator defines what a consensus validator implementation should do.
type Validator interface {
	Stake() *big.Int
	Rating() int32
	PubKey() []byte
}

// ValidatorGroupSelector defines the behaviour of a struct able to do validator group selection
type ValidatorGroupSelector interface {
	PublicKeysSelector
	LoadEligibleList(eligibleList []Validator) error
	ComputeValidatorsGroup(randomness []byte) (validatorsGroup []Validator, err error)
	ConsensusGroupSize() int
	SetConsensusGroupSize(int) error
}

// PublicKeysSelector allows retrieval of eligible validators public keys selected by a bitmap
type PublicKeysSelector interface {
	GetSelectedPublicKeys(selection []byte) (publicKeys []string, err error)
}

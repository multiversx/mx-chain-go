package mock

import (
	"math"
	"sync"
	"time"
)

// RounderMock -
type RounderMock struct {
	RoundIndex        int64
	RoundTimeStamp    time.Time
	RoundTimeDuration time.Duration
	mutRounder sync.RWMutex
}

// BeforeGenesis -
func (rndm *RounderMock) BeforeGenesis() bool {
	return false
}

// Index -
func (rndm *RounderMock) Index() int64 {
	rndm.mutRounder.RLock()
	defer rndm.mutRounder.RUnlock()
	return rndm.RoundIndex
}

// TimeDuration -
func (rndm *RounderMock) TimeDuration() time.Duration {
	rndm.mutRounder.RLock()
	defer rndm.mutRounder.RUnlock()

	return rndm.RoundTimeDuration
}

// TimeStamp -
func (rndm *RounderMock) TimeStamp() time.Time {
	rndm.mutRounder.RLock()
	defer rndm.mutRounder.RUnlock()

	return rndm.RoundTimeStamp
}

// UpdateRound -
func (rndm *RounderMock) UpdateRound(genesisRoundTimeStamp time.Time, timeStamp time.Time) {
	rndm.mutRounder.Lock()
	defer rndm.mutRounder.Lock()

	delta := timeStamp.Sub(genesisRoundTimeStamp).Nanoseconds()
	index := int64(math.Floor(float64(delta) / float64(rndm.RoundTimeDuration.Nanoseconds())))

	if rndm.RoundIndex != index {
		rndm.RoundIndex = index
		rndm.RoundTimeStamp = genesisRoundTimeStamp.Add(time.Duration(index * rndm.RoundTimeDuration.Nanoseconds()))
	}
}

// RemainingTime -
func (rndm *RounderMock) RemainingTime(_ time.Time, _ time.Duration) time.Duration {
	rndm.mutRounder.RLock()
	defer rndm.mutRounder.RUnlock()

	return rndm.RoundTimeDuration
}

// IsInterfaceNil returns true if there is no value under the interface
func (rndm *RounderMock) IsInterfaceNil() bool {
	return rndm == nil
}

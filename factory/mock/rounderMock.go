package mock

import (
	"math"
	"sync"
	"time"
)

// RoundHandlerMock -
type RoundHandlerMock struct {
	RoundIndex        int64
	RoundTimeStamp    time.Time
	RoundTimeDuration time.Duration
	mutRoundHandler   sync.RWMutex
}

// BeforeGenesis -
func (rndm *RoundHandlerMock) BeforeGenesis() bool {
	return false
}

// Index -
func (rndm *RoundHandlerMock) Index() int64 {
	rndm.mutRoundHandler.RLock()
	defer rndm.mutRoundHandler.RUnlock()
	return rndm.RoundIndex
}

// TimeDuration -
func (rndm *RoundHandlerMock) TimeDuration() time.Duration {
	rndm.mutRoundHandler.RLock()
	defer rndm.mutRoundHandler.RUnlock()

	return rndm.RoundTimeDuration
}

// TimeStamp -
func (rndm *RoundHandlerMock) TimeStamp() time.Time {
	rndm.mutRoundHandler.RLock()
	defer rndm.mutRoundHandler.RUnlock()

	return rndm.RoundTimeStamp
}

// UpdateRound -
func (rndm *RoundHandlerMock) UpdateRound(genesisRoundTimeStamp time.Time, timeStamp time.Time) {
	rndm.mutRoundHandler.Lock()
	defer rndm.mutRoundHandler.Lock()

	delta := timeStamp.Sub(genesisRoundTimeStamp).Nanoseconds()
	index := int64(math.Floor(float64(delta) / float64(rndm.RoundTimeDuration.Nanoseconds())))

	if rndm.RoundIndex != index {
		rndm.RoundIndex = index
		rndm.RoundTimeStamp = genesisRoundTimeStamp.Add(time.Duration(index * rndm.RoundTimeDuration.Nanoseconds()))
	}
}

// RemainingTime -
func (rndm *RoundHandlerMock) RemainingTime(_ time.Time, _ time.Duration) time.Duration {
	rndm.mutRoundHandler.RLock()
	defer rndm.mutRoundHandler.RUnlock()

	return rndm.RoundTimeDuration
}

// IsInterfaceNil returns true if there is no value under the interface
func (rndm *RoundHandlerMock) IsInterfaceNil() bool {
	return rndm == nil
}

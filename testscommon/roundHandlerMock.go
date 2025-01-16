package testscommon

import (
	"sync"
	"time"
)

// RoundHandlerMock -
type RoundHandlerMock struct {
	indexMut sync.RWMutex
	index    int64

	IndexCalled          func() int64
	TimeDurationCalled   func() time.Duration
	TimeStampCalled      func() time.Time
	UpdateRoundCalled    func(time.Time, time.Time)
	RemainingTimeCalled  func(startTime time.Time, maxTime time.Duration) time.Duration
	BeforeGenesisCalled  func() bool
	IncrementIndexCalled func()
}

// BeforeGenesis -
func (rndm *RoundHandlerMock) BeforeGenesis() bool {
	if rndm.BeforeGenesisCalled != nil {
		return rndm.BeforeGenesisCalled()
	}
	return false
}

// Index -
func (rndm *RoundHandlerMock) Index() int64 {
	if rndm.IndexCalled != nil {
		return rndm.IndexCalled()
	}

	rndm.indexMut.RLock()
	defer rndm.indexMut.RUnlock()

	return rndm.index
}

// TimeDuration -
func (rndm *RoundHandlerMock) TimeDuration() time.Duration {
	if rndm.TimeDurationCalled != nil {
		return rndm.TimeDurationCalled()
	}

	return 4000 * time.Millisecond
}

// TimeStamp -
func (rndm *RoundHandlerMock) TimeStamp() time.Time {
	if rndm.TimeStampCalled != nil {
		return rndm.TimeStampCalled()
	}

	return time.Unix(0, 0)
}

// UpdateRound -
func (rndm *RoundHandlerMock) UpdateRound(genesisRoundTimeStamp time.Time, timeStamp time.Time) {
	if rndm.UpdateRoundCalled != nil {
		rndm.UpdateRoundCalled(genesisRoundTimeStamp, timeStamp)
		return
	}

	rndm.indexMut.Lock()
	rndm.index++
	rndm.indexMut.Unlock()
}

// RemainingTime -
func (rndm *RoundHandlerMock) RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration {
	if rndm.RemainingTimeCalled != nil {
		return rndm.RemainingTimeCalled(startTime, maxTime)
	}

	return 4000 * time.Millisecond
}

// IncrementIndex -
func (rndm *RoundHandlerMock) IncrementIndex() {
	if rndm.IncrementIndexCalled != nil {
		rndm.IncrementIndexCalled()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (rndm *RoundHandlerMock) IsInterfaceNil() bool {
	return rndm == nil
}

package mock

import "time"

// RoundHandlerMock -
type RoundHandlerMock struct {
	IndexField          int64
	TimeStampField      time.Time
	TimeDurationField   time.Duration
	RemainingTimeField  time.Duration
	BeforeGenesisCalled func() bool
}

// BeforeGenesis -
func (rndm *RoundHandlerMock) BeforeGenesis() bool {
	if rndm.BeforeGenesisCalled != nil {
		return rndm.BeforeGenesisCalled()
	}
	return false
}

// Index -
func (rm *RoundHandlerMock) Index() int64 {
	return rm.IndexField
}

// UpdateRound -
func (rm *RoundHandlerMock) UpdateRound(time.Time, time.Time) {
}

// TimeStamp -
func (rm *RoundHandlerMock) TimeStamp() time.Time {
	return rm.TimeStampField
}

// TimeDuration -
func (rm *RoundHandlerMock) TimeDuration() time.Duration {
	if rm.TimeDurationField.Seconds() == 0 {
		return time.Second
	}

	return rm.TimeDurationField
}

// RemainingTime -
func (rm *RoundHandlerMock) RemainingTime(_ time.Time, _ time.Duration) time.Duration {
	return rm.RemainingTimeField
}

// IsInterfaceNil -
func (rm *RoundHandlerMock) IsInterfaceNil() bool {
	return rm == nil
}

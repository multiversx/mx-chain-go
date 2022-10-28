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
func (rndm *RoundHandlerMock) Index() int64 {
	return rndm.IndexField
}

// UpdateRound -
func (rndm *RoundHandlerMock) UpdateRound(time.Time, time.Time) {
}

// TimeStamp -
func (rndm *RoundHandlerMock) TimeStamp() time.Time {
	return rndm.TimeStampField
}

// TimeDuration -
func (rndm *RoundHandlerMock) TimeDuration() time.Duration {
	if rndm.TimeDurationField.Seconds() == 0 {
		return time.Second
	}

	return rndm.TimeDurationField
}

// RemainingTime -
func (rndm *RoundHandlerMock) RemainingTime(_ time.Time, _ time.Duration) time.Duration {
	return rndm.RemainingTimeField
}

// IsInterfaceNil -
func (rndm *RoundHandlerMock) IsInterfaceNil() bool {
	return rndm == nil
}

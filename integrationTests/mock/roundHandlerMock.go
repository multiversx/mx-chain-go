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
func (mock *RoundHandlerMock) BeforeGenesis() bool {
	if mock.BeforeGenesisCalled != nil {
		return mock.BeforeGenesisCalled()
	}
	return false
}

// Index -
func (mock *RoundHandlerMock) Index() int64 {
	return mock.IndexField
}

// UpdateRound -
func (mock *RoundHandlerMock) UpdateRound(time.Time, time.Time) {
}

// TimeStamp -
func (mock *RoundHandlerMock) TimeStamp() time.Time {
	return mock.TimeStampField
}

// TimeDuration -
func (mock *RoundHandlerMock) TimeDuration() time.Duration {
	if mock.TimeDurationField.Seconds() == 0 {
		return time.Second
	}

	return mock.TimeDurationField
}

// RemainingTime -
func (mock *RoundHandlerMock) RemainingTime(_ time.Time, _ time.Duration) time.Duration {
	return mock.RemainingTimeField
}

// IsInterfaceNil -
func (mock *RoundHandlerMock) IsInterfaceNil() bool {
	return mock == nil
}

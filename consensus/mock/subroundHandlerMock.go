package mock

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-go/consensus"
)

// SubroundHandlerMock -
type SubroundHandlerMock struct {
	DoWorkCalled                 func(roundHandler consensus.RoundHandler) bool
	PreviousCalled               func() int
	NextCalled                   func() int
	CurrentCalled                func() int
	StartTimeCalled              func() int64
	EndTimeCalled                func() int64
	SetBaseDurationCalled        func(baseDuration time.Duration)
	SetStartTimePercentageCalled func(startTimePercent float64)
	SetEndTimePercentageCalled   func(endTimePercent float64)

	SetSignatureSubroundEndTimePercentageCalled func(percent float64)

	NameCalled             func() string
	JobCalled              func() bool
	CheckCalled            func() bool
	ConsensusChannelCalled func() chan bool
}

// DoWork -
func (srm *SubroundHandlerMock) DoWork(_ context.Context, roundHandler consensus.RoundHandler) bool {
	return srm.DoWorkCalled(roundHandler)
}

// Previous -
func (srm *SubroundHandlerMock) Previous() int {
	return srm.PreviousCalled()
}

// Next -
func (srm *SubroundHandlerMock) Next() int {
	return srm.NextCalled()
}

// Current -
func (srm *SubroundHandlerMock) Current() int {
	return srm.CurrentCalled()
}

// StartTime -
func (srm *SubroundHandlerMock) StartTime() int64 {
	return srm.StartTimeCalled()
}

// EndTime -
func (srm *SubroundHandlerMock) EndTime() int64 {
	return srm.EndTimeCalled()
}

// SetBaseDuration -
func (srm *SubroundHandlerMock) SetBaseDuration(baseDuration time.Duration) {
	if srm.SetBaseDurationCalled != nil {
		srm.SetBaseDurationCalled(baseDuration)
	}
}

// SetStartTimePercentage -
func (srm *SubroundHandlerMock) SetStartTimePercentage(startTimePercent float64) {
	if srm.SetStartTimePercentageCalled != nil {
		srm.SetStartTimePercentageCalled(startTimePercent)
	}
}

// SetEndTimePercentage -
func (srm *SubroundHandlerMock) SetEndTimePercentage(endTimePercent float64) {
	if srm.SetEndTimePercentageCalled != nil {
		srm.SetEndTimePercentageCalled(endTimePercent)
	}
}

// SetSignatureSubroundEndTimePercentage -
func (srm *SubroundHandlerMock) SetSignatureSubroundEndTimePercentage(percent float64) {
	if srm.SetSignatureSubroundEndTimePercentageCalled != nil {
		srm.SetSignatureSubroundEndTimePercentageCalled(percent)
	}
}

// Name -
func (srm *SubroundHandlerMock) Name() string {
	return srm.NameCalled()
}

// ConsensusChannel -
func (srm *SubroundHandlerMock) ConsensusChannel() chan bool {
	return srm.ConsensusChannelCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (srm *SubroundHandlerMock) IsInterfaceNil() bool {
	return srm == nil
}

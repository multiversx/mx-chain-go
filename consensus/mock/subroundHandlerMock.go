package mock

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-go/consensus"
)

// SubroundHandlerMock -
type SubroundHandlerMock struct {
	DoWorkCalled              func(roundHandler consensus.RoundHandler) bool
	PreviousCalled            func() int
	NextCalled                func() int
	CurrentCalled             func() int
	StartTimeCalled           func() int64
	EndTimeCalled             func() int64
	SetBaseDurationCalled     func(baseDuration time.Duration)
	SetTimingPercentageCalled func(startTimePercent float64, endTimePercent float64)

	SetSignatureSubroundEndTimePercentageCalled func(percent float64)
	SetProcessingThresholdPercentCalled         func(percent int)

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

// SetTimingPercentage -
func (srm *SubroundHandlerMock) SetTimingPercentage(startTimePercent float64, endTimePercent float64) {
	if srm.SetTimingPercentageCalled != nil {
		srm.SetTimingPercentageCalled(startTimePercent, endTimePercent)
	}
}

// SetSignatureSubroundEndTimePercentage -
func (srm *SubroundHandlerMock) SetSignatureSubroundEndTimePercentage(percent float64) {
	if srm.SetSignatureSubroundEndTimePercentageCalled != nil {
		srm.SetSignatureSubroundEndTimePercentageCalled(percent)
	}
}

// SetProcessingThresholdPercent -
func (srm *SubroundHandlerMock) SetProcessingThresholdPercent(percent int) {
	if srm.SetProcessingThresholdPercentCalled != nil {
		srm.SetProcessingThresholdPercentCalled(percent)
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

package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
)

type SubroundHandlerMock struct {
	DoWorkCalled    func(rounder consensus.Rounder) bool
	PreviousCalled  func() int
	NextCalled      func() int
	CurrentCalled   func() int
	StartTimeCalled func() int64
	EndTimeCalled   func() int64
	NameCalled      func() string
	JobCalled       func() bool
	CheckCalled     func() bool
}

func (srm *SubroundHandlerMock) DoWork(rounder consensus.Rounder) bool {
	return srm.DoWorkCalled(rounder)
}

func (srm *SubroundHandlerMock) Previous() int {
	return srm.PreviousCalled()
}

func (srm *SubroundHandlerMock) Next() int {
	return srm.NextCalled()
}

func (srm *SubroundHandlerMock) Current() int {
	return srm.CurrentCalled()
}

func (srm *SubroundHandlerMock) StartTime() int64 {
	return srm.StartTimeCalled()
}

func (srm *SubroundHandlerMock) EndTime() int64 {
	return srm.EndTimeCalled()
}

func (srm *SubroundHandlerMock) Name() string {
	return srm.NameCalled()
}

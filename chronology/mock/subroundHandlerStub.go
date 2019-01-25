package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SubroundHandlerStub struct {
	DoWorkCalled  func(func() chronology.SubroundId, func() bool) bool
	NextCalled    func() chronology.SubroundId
	CurrentCalled func() chronology.SubroundId
	EndTimeCalled func() int64
	NameCalled    func() string
	CheckCalled   func() bool
}

func (shs *SubroundHandlerStub) DoWork(handler1 func() chronology.SubroundId, handler2 func() bool) bool {
	return shs.DoWorkCalled(handler1, handler2)
}

func (shs *SubroundHandlerStub) Next() chronology.SubroundId {
	return shs.NextCalled()
}

func (shs *SubroundHandlerStub) Current() chronology.SubroundId {
	return shs.CurrentCalled()
}

func (shs *SubroundHandlerStub) EndTime() int64 {
	return shs.EndTimeCalled()
}

func (shs *SubroundHandlerStub) Name() string {
	return shs.NameCalled()
}

func (shs *SubroundHandlerStub) Check() bool {
	return shs.CheckCalled()
}

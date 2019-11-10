package mock

import "github.com/ElrondNetwork/elrond-go/data"

type EndOfEpochTriggerStub struct {
	ForceEndOfEpochCalled func() error
	IsEndOfEpochCalled    func() bool
	EpochCalled           func() uint32
	ReceivedHeaderCalled  func(handler data.HeaderHandler)
}

func (e *EndOfEpochTriggerStub) ForceEndOfEpoch() error {
	if e.ForceEndOfEpochCalled != nil {
		return e.ForceEndOfEpochCalled()
	}
	return nil
}

func (e *EndOfEpochTriggerStub) IsEndOfEpoch() bool {
	if e.IsEndOfEpochCalled != nil {
		return e.IsEndOfEpochCalled()
	}
	return false
}

func (e *EndOfEpochTriggerStub) Epoch() uint32 {
	if e.EpochCalled != nil {
		return e.EpochCalled()
	}
	return 0
}

func (e *EndOfEpochTriggerStub) ReceivedHeader(header data.HeaderHandler) {
	if e.ReceivedHeaderCalled != nil {
		e.ReceivedHeaderCalled(header)
	}
}

func (e *EndOfEpochTriggerStub) IsInterfaceNil() bool {
	return e == nil
}

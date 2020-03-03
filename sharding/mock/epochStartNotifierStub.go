package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

type EpochStartNotifierStub struct {
	RegisterHandlerCalled   func(handler epochStart.EpochStartHandler)
	UnregisterHandlerCalled func(handler epochStart.EpochStartHandler)
	NotifyAllPrepareCalled  func(hdr data.HeaderHandler)
	NotifyAllCalled         func(hdr data.HeaderHandler)
}

func (esnm *EpochStartNotifierStub) RegisterHandler(handler epochStart.EpochStartHandler) {
	if esnm.RegisterHandlerCalled != nil {
		esnm.RegisterHandlerCalled(handler)
	}
}

func (esnm *EpochStartNotifierStub) UnregisterHandler(handler epochStart.EpochStartHandler) {
	if esnm.UnregisterHandlerCalled != nil {
		esnm.UnregisterHandlerCalled(handler)
	}
}

func (esnm *EpochStartNotifierStub) NotifyAllPrepare(metaHeader data.HeaderHandler) {
	if esnm.NotifyAllPrepareCalled != nil {
		esnm.NotifyAllPrepareCalled(metaHeader)
	}
}

func (esnm *EpochStartNotifierStub) NotifyAll(hdr data.HeaderHandler) {
	if esnm.NotifyAllCalled != nil {
		esnm.NotifyAllCalled(hdr)
	}
}

func (esnm *EpochStartNotifierStub) IsInterfaceNil() bool {
	return esnm == nil
}

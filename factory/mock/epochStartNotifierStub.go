package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// EpochStartNotifierStub -
type EpochStartNotifierStub struct {
	RegisterHandlerCalled            func(handler epochStart.ActionHandler)
	UnregisterHandlerCalled          func(handler epochStart.ActionHandler)
	NotifyAllPrepareCalled           func(hdr data.HeaderHandler, body data.BodyHandler)
	NotifyAllCalled                  func(hdr data.HeaderHandler)
	NotifyEpochChangeConfirmedCalled func(epoch uint32)
}

// RegisterHandler -
func (esnm *EpochStartNotifierStub) RegisterHandler(handler epochStart.ActionHandler) {
	if esnm.RegisterHandlerCalled != nil {
		esnm.RegisterHandlerCalled(handler)
	}
}

// UnregisterHandler -
func (esnm *EpochStartNotifierStub) UnregisterHandler(handler epochStart.ActionHandler) {
	if esnm.UnregisterHandlerCalled != nil {
		esnm.UnregisterHandlerCalled(handler)
	}
}

// NotifyAllPrepare -
func (esnm *EpochStartNotifierStub) NotifyAllPrepare(metaHdr data.HeaderHandler, body data.BodyHandler) {
	if esnm.NotifyAllPrepareCalled != nil {
		esnm.NotifyAllPrepareCalled(metaHdr, body)
	}
}

// NotifyAll -
func (esnm *EpochStartNotifierStub) NotifyAll(hdr data.HeaderHandler) {
	if esnm.NotifyAllCalled != nil {
		esnm.NotifyAllCalled(hdr)
	}
}

// NotifyEpochChangeConfirmed -
func (esnm *EpochStartNotifierStub) NotifyEpochChangeConfirmed(epoch uint32) {
	if esnm.NotifyEpochChangeConfirmedCalled != nil {
		esnm.NotifyEpochChangeConfirmedCalled(epoch)
	}
}

// IsInterfaceNil -
func (esnm *EpochStartNotifierStub) IsInterfaceNil() bool {
	return esnm == nil
}

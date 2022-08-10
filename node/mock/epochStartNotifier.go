package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
)

// EpochStartNotifierStub -
type EpochStartNotifierStub struct {
	RegisterHandlerCalled            func(handler core.EpochStartActionHandler)
	UnregisterHandlerCalled          func(handler core.EpochStartActionHandler)
	NotifyAllCalled                  func(hdr data.HeaderHandler)
	NotifyAllPrepareCalled           func(hdr data.HeaderHandler, body data.BodyHandler)
	NotifyEpochChangeConfirmedCalled func(epoch uint32)
	epochStartHdls                   []core.EpochStartActionHandler
}

// RegisterHandler -
func (esnm *EpochStartNotifierStub) RegisterHandler(handler core.EpochStartActionHandler) {
	if esnm.RegisterHandlerCalled != nil {
		esnm.RegisterHandlerCalled(handler)
	}

	esnm.epochStartHdls = append(esnm.epochStartHdls, handler)
}

// UnregisterHandler -
func (esnm *EpochStartNotifierStub) UnregisterHandler(handler core.EpochStartActionHandler) {
	if esnm.UnregisterHandlerCalled != nil {
		esnm.UnregisterHandlerCalled(handler)
	}

	for i, hdl := range esnm.epochStartHdls {
		if hdl == handler {
			esnm.epochStartHdls = append(esnm.epochStartHdls[:i], esnm.epochStartHdls[i+1:]...)
			break
		}
	}
}

// NotifyAllPrepare -
func (esnm *EpochStartNotifierStub) NotifyAllPrepare(metaHdr data.HeaderHandler, body data.BodyHandler) {
	if esnm.NotifyAllPrepareCalled != nil {
		esnm.NotifyAllPrepareCalled(metaHdr, body)
	}

	for _, hdl := range esnm.epochStartHdls {
		hdl.EpochStartPrepare(metaHdr, body)
	}
}

// NotifyAll -
func (esnm *EpochStartNotifierStub) NotifyAll(hdr data.HeaderHandler) {
	if esnm.NotifyAllCalled != nil {
		esnm.NotifyAllCalled(hdr)
	}

	for _, hdl := range esnm.epochStartHdls {
		hdl.EpochStartAction(hdr)
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

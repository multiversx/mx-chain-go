package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// EpochStartNotifierStub -
type EpochStartNotifierStub struct {
	RegisterHandlerCalled   func(handler epochStart.ActionHandler)
	UnregisterHandlerCalled func(handler epochStart.ActionHandler)
	NotifyAllCalled         func(hdr data.HeaderHandler)
	NotifyAllPrepareCalled  func(hdr data.HeaderHandler, body data.BodyHandler)
	epochStartHdls          []epochStart.ActionHandler
}

// RegisterHandler -
func (esns *EpochStartNotifierStub) RegisterHandler(handler epochStart.ActionHandler) {
	if esns.RegisterHandlerCalled != nil {
		esns.RegisterHandlerCalled(handler)
	}

	esns.epochStartHdls = append(esns.epochStartHdls, handler)
}

// UnregisterHandler -
func (esns *EpochStartNotifierStub) UnregisterHandler(handler epochStart.ActionHandler) {
	if esns.UnregisterHandlerCalled != nil {
		esns.UnregisterHandlerCalled(handler)
	}

	for i, hdl := range esns.epochStartHdls {
		if hdl == handler {
			esns.epochStartHdls = append(esns.epochStartHdls[:i], esns.epochStartHdls[i+1:]...)
			break
		}
	}
}

// NotifyAllPrepare -
func (esns *EpochStartNotifierStub) NotifyAllPrepare(metaHdr data.HeaderHandler, body data.BodyHandler) {
	if esns.NotifyAllPrepareCalled != nil {
		esns.NotifyAllPrepareCalled(metaHdr, body)
	}

	for _, hdl := range esns.epochStartHdls {
		hdl.EpochStartPrepare(metaHdr, body)
	}
}

// NotifyAll -
func (esns *EpochStartNotifierStub) NotifyAll(hdr data.HeaderHandler) {
	if esns.NotifyAllCalled != nil {
		esns.NotifyAllCalled(hdr)
	}

	for _, hdl := range esns.epochStartHdls {
		hdl.EpochStartAction(hdr)
	}
}

// IsInterfaceNil -
func (esns *EpochStartNotifierStub) IsInterfaceNil() bool {
	return esns == nil
}

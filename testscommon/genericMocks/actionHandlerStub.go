package genericMocks

import "github.com/ElrondNetwork/elrond-go-core/data"

// ActionHandlerStub -
type ActionHandlerStub struct {
	EpochStartActionCalled  func(hdr data.HeaderHandler)
	EpochStartPrepareCalled func(metaHdr data.HeaderHandler, body data.BodyHandler)
	NotifyOrderCalled       func() uint32
}

// EpochStartAction -
func (ahs *ActionHandlerStub) Action(hdr data.HeaderHandler) {
	if ahs.EpochStartActionCalled != nil {
		ahs.EpochStartActionCalled(hdr)
	}
}

// EpochStartPrepare -
func (ahs *ActionHandlerStub) Prepare(metaHdr data.HeaderHandler, body data.BodyHandler) {
	if ahs.EpochStartPrepareCalled != nil {
		ahs.EpochStartPrepareCalled(metaHdr, body)
	}
}

// NotifyOrder -
func (ahs *ActionHandlerStub) NotifyOrder() uint32 {
	if ahs.NotifyOrderCalled != nil {
		return ahs.NotifyOrderCalled()
	}

	return 0
}

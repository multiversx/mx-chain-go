package genericMocks

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// ActionHandlerStub -
type ActionHandlerStub struct {
	EpochStartActionCalled  func(hdr data.HeaderHandler)
	EpochStartPrepareCalled func(metaHdr data.HeaderHandler, body data.BodyHandler, validatorInfoCacher epochStart.ValidatorInfoCacher)
	NotifyOrderCalled       func() uint32
}

// EpochStartAction -
func (ahs *ActionHandlerStub) EpochStartAction(hdr data.HeaderHandler) {
	if ahs.EpochStartActionCalled != nil {
		ahs.EpochStartActionCalled(hdr)
	}
}

// EpochStartPrepare -
func (ahs *ActionHandlerStub) EpochStartPrepare(metaHdr data.HeaderHandler, body data.BodyHandler, validatorInfoCacher epochStart.ValidatorInfoCacher) {
	if ahs.EpochStartPrepareCalled != nil {
		ahs.EpochStartPrepareCalled(metaHdr, body, validatorInfoCacher)
	}
}

// NotifyOrder -
func (ahs *ActionHandlerStub) NotifyOrder() uint32 {
	if ahs.NotifyOrderCalled != nil {
		return ahs.NotifyOrderCalled()
	}

	return 0
}

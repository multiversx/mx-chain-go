package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// EpochStartNotifierStub -
type EpochStartNotifierStub struct {
	NotifyAllCalled                  func(hdr data.HeaderHandler)
	NotifyAllPrepareCalled           func(hdr data.HeaderHandler, body data.BodyHandler, validatorInfoCacher epochStart.ValidatorInfoCacher)
	NotifyEpochChangeConfirmedCalled func(epoch uint32)
}

// NotifyEpochChangeConfirmed -
func (esnm *EpochStartNotifierStub) NotifyEpochChangeConfirmed(epoch uint32) {
	if esnm.NotifyEpochChangeConfirmedCalled != nil {
		esnm.NotifyEpochChangeConfirmedCalled(epoch)
	}
}

// NotifyAll -
func (esnm *EpochStartNotifierStub) NotifyAll(hdr data.HeaderHandler) {
	if esnm.NotifyAllCalled != nil {
		esnm.NotifyAllCalled(hdr)
	}
}

// NotifyAllPrepare -
func (esnm *EpochStartNotifierStub) NotifyAllPrepare(metaHdr data.HeaderHandler, body data.BodyHandler, validatorInfoCacher epochStart.ValidatorInfoCacher) {
	if esnm.NotifyAllPrepareCalled != nil {
		esnm.NotifyAllPrepareCalled(metaHdr, body, validatorInfoCacher)
	}
}

// IsInterfaceNil -
func (esnm *EpochStartNotifierStub) IsInterfaceNil() bool {
	return esnm == nil
}

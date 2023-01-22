package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// EpochStartNotifierStub -
type EpochStartNotifierStub struct {
	NotifyAllCalled                  func(hdr data.HeaderHandler)
	NotifyAllPrepareCalled           func(hdr data.HeaderHandler, body data.BodyHandler)
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
func (esnm *EpochStartNotifierStub) NotifyAllPrepare(metaHdr data.HeaderHandler, body data.BodyHandler) {
	if esnm.NotifyAllPrepareCalled != nil {
		esnm.NotifyAllPrepareCalled(metaHdr, body)
	}
}

// IsInterfaceNil -
func (esnm *EpochStartNotifierStub) IsInterfaceNil() bool {
	return esnm == nil
}

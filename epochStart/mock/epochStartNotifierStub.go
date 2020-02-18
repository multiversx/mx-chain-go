package mock

import "github.com/ElrondNetwork/elrond-go/data"

// EpochStartNotifierStub -
type EpochStartNotifierStub struct {
	NotifyAllCalled func(hdr data.HeaderHandler)
}

// NotifyAll -
func (esnm *EpochStartNotifierStub) NotifyAll(hdr data.HeaderHandler) {
	if esnm.NotifyAllCalled != nil {
		esnm.NotifyAllCalled(hdr)
	}
}

// IsInterfaceNil -
func (esnm *EpochStartNotifierStub) IsInterfaceNil() bool {
	return esnm == nil
}

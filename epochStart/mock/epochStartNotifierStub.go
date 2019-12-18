package mock

import "github.com/ElrondNetwork/elrond-go/data"

type EpochStartNotifierStub struct {
	NotifyAllCalled func(hdr data.HeaderHandler)
}

func (esnm *EpochStartNotifierStub) NotifyAll(hdr data.HeaderHandler) {
	esnm.NotifyAllCalled(hdr)
}

func (esnm *EpochStartNotifierStub) IsInterfaceNil() bool {
	return esnm == nil
}

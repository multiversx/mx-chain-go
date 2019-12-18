package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
)

type EpochStartNotifierStub struct {
	RegisterHandlerCalled   func(handler notifier.SubscribeFunctionHandler)
	UnregisterHandlerCalled func(handler notifier.SubscribeFunctionHandler)
	NotifyAllCalled         func(hdr data.HeaderHandler)
}

func (esnm *EpochStartNotifierStub) RegisterHandler(handler notifier.SubscribeFunctionHandler) {
	if esnm.RegisterHandlerCalled != nil {
		esnm.RegisterHandlerCalled(handler)
	}
}

func (esnm *EpochStartNotifierStub) UnregisterHandler(handler notifier.SubscribeFunctionHandler) {
	if esnm.UnregisterHandlerCalled != nil {
		esnm.UnregisterHandlerCalled(handler)
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

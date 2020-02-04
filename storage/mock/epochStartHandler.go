package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
)

// EpochStartNotifierStub -
type EpochStartNotifierStub struct {
	RegisterHandlerCalled   func(handler notifier.SubscribeFunctionHandler)
	UnregisterHandlerCalled func(handler notifier.SubscribeFunctionHandler)
	NotifyAllCalled         func(hdr data.HeaderHandler)
}

// RegisterHandler -
func (esnm *EpochStartNotifierStub) RegisterHandler(handler notifier.SubscribeFunctionHandler) {
	if esnm.RegisterHandlerCalled != nil {
		esnm.RegisterHandlerCalled(handler)
	}
}

// UnregisterHandler -
func (esnm *EpochStartNotifierStub) UnregisterHandler(handler notifier.SubscribeFunctionHandler) {
	if esnm.UnregisterHandlerCalled != nil {
		esnm.UnregisterHandlerCalled(handler)
	}
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

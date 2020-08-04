package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
)

// EpochNotifierStub -
type EpochNotifierStub struct {
	CurrentEpochCalled          func() uint32
	RegisterNotifyHandlerCalled func(handler core.EpochNotifiedHandler)
}

// RegisterNotifyHandler -
func (ens *EpochNotifierStub) RegisterNotifyHandler(handler core.EpochNotifiedHandler) {
	if ens.RegisterNotifyHandlerCalled != nil {
		ens.RegisterNotifyHandlerCalled(handler)
	} else {
		if !check.IfNil(handler) {
			handler.EpochConfirmed(0)
		}
	}
}

// CurrentEpoch -
func (ens *EpochNotifierStub) CurrentEpoch() uint32 {
	if ens.CurrentEpochCalled != nil {
		return ens.CurrentEpochCalled()
	}

	return 0
}

// IsInterfaceNil -
func (ens *EpochNotifierStub) IsInterfaceNil() bool {
	return ens == nil
}

package mock

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// EpochNotifierStub -
type EpochNotifierStub struct {
	CheckEpochCalled            func(header data.HeaderHandler)
	CurrentEpochCalled          func() uint32
	RegisterNotifyHandlerCalled func(handler vmcommon.EpochSubscriberHandler)
}

// CheckEpoch -
func (ens *EpochNotifierStub) CheckEpoch(header data.HeaderHandler) {
	if ens.CheckEpochCalled != nil {
		ens.CheckEpochCalled(header)
	}
}

// RegisterNotifyHandler -
func (ens *EpochNotifierStub) RegisterNotifyHandler(handler vmcommon.EpochSubscriberHandler) {
	if ens.RegisterNotifyHandlerCalled != nil {
		ens.RegisterNotifyHandlerCalled(handler)
	} else {
		if !check.IfNil(handler) {
			handler.EpochConfirmed(0, 0)
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

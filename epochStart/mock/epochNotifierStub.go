package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// EpochNotifierStub -
type EpochNotifierStub struct {
	NewEpochCalled              func(epoch uint32)
	CheckEpochCalled            func(header data.HeaderHandler)
	CurrentEpochCalled          func() uint32
	RegisterNotifyHandlerCalled func(handler vmcommon.EpochSubscriberHandler)
	RegisterHandlerCalled       func(handler epochStart.ActionHandler)
}

// RegisterHandler -
func (ens *EpochNotifierStub) RegisterHandler(handler epochStart.ActionHandler) {
	if ens.RegisterHandlerCalled != nil {
		ens.RegisterHandlerCalled(handler)
	}
}

// NewEpoch -
func (ens *EpochNotifierStub) NewEpoch(epoch uint32) {
	if ens.NewEpochCalled != nil {
		ens.NewEpochCalled(epoch)
	}
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

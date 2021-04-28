package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// EpochNotifierStub -
type EpochNotifierStub struct {
	NewEpochCalled              func(epoch uint32, timestamp uint64)
	RegisterHandlerCalled       func(handler epochStart.ActionHandler)
	CheckEpochCalled            func(header data.HeaderHandler)
	CurrentEpochCalled          func() uint32
	RegisterNotifyHandlerCalled func(handler core.EpochSubscriberHandler)
}

// NewEpoch -
func (ens *EpochNotifierStub) NewEpoch(epoch uint32, timestamp uint64) {
	if ens.NewEpochCalled != nil {
		ens.NewEpochCalled(epoch, timestamp)
	}
}

// RegisterHandler -
func (ens *EpochNotifierStub) RegisterHandler(handler epochStart.ActionHandler) {
	if ens.RegisterHandlerCalled != nil {
		ens.RegisterHandlerCalled(handler)
	}
}

// CheckEpoch -
func (ens *EpochNotifierStub) CheckEpoch(header data.HeaderHandler) {
	if ens.CheckEpochCalled != nil {
		ens.CheckEpochCalled(header)
	}
}

// RegisterNotifyHandler -
func (ens *EpochNotifierStub) RegisterNotifyHandler(handler core.EpochSubscriberHandler) {
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

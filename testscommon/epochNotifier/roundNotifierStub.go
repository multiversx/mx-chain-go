package epochNotifier

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// RoundNotifierStub -
type RoundNotifierStub struct {
	NewRoundCalled              func(Round uint32, timestamp uint64)
	CheckRoundCalled            func(header data.HeaderHandler)
	CurrentRoundCalled          func() uint64
	RegisterNotifyHandlerCalled func(handler vmcommon.RoundSubscriberHandler)
}

// NewRound -
func (ens *RoundNotifierStub) NewRound(Round uint32, timestamp uint64) {
	if ens.NewRoundCalled != nil {
		ens.NewRoundCalled(Round, timestamp)
	}
}

// CheckRound -
func (ens *RoundNotifierStub) CheckRound(header data.HeaderHandler) {
	if ens.CheckRoundCalled != nil {
		ens.CheckRoundCalled(header)
	}
}

// RegisterNotifyHandler -
func (ens *RoundNotifierStub) RegisterNotifyHandler(handler vmcommon.RoundSubscriberHandler) {
	if ens.RegisterNotifyHandlerCalled != nil {
		ens.RegisterNotifyHandlerCalled(handler)
	} else {
		if !check.IfNil(handler) {
			handler.RoundConfirmed(0, 0)
		}
	}
}

// CurrentRound -
func (ens *RoundNotifierStub) CurrentRound() uint64 {
	if ens.CurrentRoundCalled != nil {
		return ens.CurrentRoundCalled()
	}

	return 0
}

// IsInterfaceNil -
func (ens *RoundNotifierStub) IsInterfaceNil() bool {
	return ens == nil
}

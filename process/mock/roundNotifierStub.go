package mock

import "github.com/ElrondNetwork/elrond-go/process"

// RoundNotifierStub -
type RoundNotifierStub struct {
	CheckRoundCalled func(round uint64)
}

// RegisterNotifyHandler -
func (rns *RoundNotifierStub) RegisterNotifyHandler(process.RoundSubscriberHandler) {
}

// CheckRound -
func (rns *RoundNotifierStub) CheckRound(round uint64) {
	if rns.CheckRoundCalled != nil {
		rns.CheckRoundCalled(round)
	}
}

// IsInterfaceNil -
func (rns *RoundNotifierStub) IsInterfaceNil() bool {
	return rns == nil
}

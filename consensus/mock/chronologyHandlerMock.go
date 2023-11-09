package mock

import (
	"github.com/multiversx/mx-chain-go/consensus"
)

// ChronologyHandlerMock -
type ChronologyHandlerMock struct {
	AddSubroundCalled        func(consensus.SubroundHandler)
	RemoveAllSubroundsCalled func()
	StartRoundCalled         func()
	EpochCalled              func() uint32
}

// Epoch -
func (chrm *ChronologyHandlerMock) Epoch() uint32 {
	if chrm.EpochCalled != nil {
		return chrm.EpochCalled()
	}
	return 0
}

// AddSubround -
func (chrm *ChronologyHandlerMock) AddSubround(subroundHandler consensus.SubroundHandler) {
	if chrm.AddSubroundCalled != nil {
		chrm.AddSubroundCalled(subroundHandler)
	}
}

// RemoveAllSubrounds -
func (chrm *ChronologyHandlerMock) RemoveAllSubrounds() {
	if chrm.RemoveAllSubroundsCalled != nil {
		chrm.RemoveAllSubroundsCalled()
	}
}

// StartRounds -
func (chrm *ChronologyHandlerMock) StartRounds() {
	if chrm.StartRoundCalled != nil {
		chrm.StartRoundCalled()
	}
}

// Close -
func (chrm *ChronologyHandlerMock) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (chrm *ChronologyHandlerMock) IsInterfaceNil() bool {
	return chrm == nil
}

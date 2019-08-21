package mock

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
)

type ChronologyHandlerMock struct {
	AddSubroundCalled        func(consensus.SubroundHandler)
	RemoveAllSubroundsCalled func()
	StartRoundCalled         func()
}

func (chrm *ChronologyHandlerMock) AddSubround(subroundHandler consensus.SubroundHandler) {
	if chrm.AddSubroundCalled != nil {
		chrm.AddSubroundCalled(subroundHandler)
	}
}

func (chrm *ChronologyHandlerMock) RemoveAllSubrounds() {
	if chrm.RemoveAllSubroundsCalled != nil {
		chrm.RemoveAllSubroundsCalled()
	}
}

func (chrm *ChronologyHandlerMock) StartRounds() {
	if chrm.StartRoundCalled != nil {
		chrm.StartRoundCalled()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (chrm *ChronologyHandlerMock) IsInterfaceNil() bool {
	if chrm == nil {
		return true
	}
	return false
}

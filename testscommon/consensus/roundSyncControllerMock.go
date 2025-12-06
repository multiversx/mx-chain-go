package consensus

// RoundSyncControllerMock -
type RoundSyncControllerMock struct {
	AddOutOfRangeRoundCalled func(round uint64)
}

// AddOutOfRangeRound -
func (mock *RoundSyncControllerMock) AddOutOfRangeRound(round uint64) {
	if mock.AddOutOfRangeRoundCalled != nil {
		mock.AddOutOfRangeRoundCalled(round)
	}
}

// IsInterfaceNil -
func (mock *RoundSyncControllerMock) IsInterfaceNil() bool {
	return mock == nil
}

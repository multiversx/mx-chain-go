package consensus

// RoundSyncControllerMock -
type RoundSyncControllerMock struct {
	AddOutOfRangeRoundCalled func(round uint64, hash string)
}

// AddOutOfRangeRound -
func (mock *RoundSyncControllerMock) AddOutOfRangeRound(round uint64, hash string) {
	if mock.AddOutOfRangeRoundCalled != nil {
		mock.AddOutOfRangeRoundCalled(round, hash)
	}
}

// IsInterfaceNil -
func (mock *RoundSyncControllerMock) IsInterfaceNil() bool {
	return mock == nil
}

package consensus

// RoundSyncControllerMock -
type RoundSyncControllerMock struct {
	AddOutOfRangeRoundCalled         func(round uint64, hash string)
	AddLeaderRoundAsOutOfRangeCalled func(round uint64, hash string)
}

// AddOutOfRangeRound -
func (mock *RoundSyncControllerMock) AddOutOfRangeRound(round uint64, hash string) {
	if mock.AddOutOfRangeRoundCalled != nil {
		mock.AddOutOfRangeRoundCalled(round, hash)
	}
}

// AddLeaderRoundAsOutOfRange -
func (mock *RoundSyncControllerMock) AddLeaderRoundAsOutOfRange(round uint64, hash string) {
	if mock.AddLeaderRoundAsOutOfRangeCalled != nil {
		mock.AddLeaderRoundAsOutOfRangeCalled(round, hash)
	}
}

// IsInterfaceNil -
func (mock *RoundSyncControllerMock) IsInterfaceNil() bool {
	return mock == nil
}

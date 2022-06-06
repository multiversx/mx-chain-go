package testscommon

// EnableRoundsHandlerMock -
type EnableRoundsHandlerMock struct {
	CheckRoundCalled                       func(round uint64)
	IsCheckValueOnExecByCallerEnabledValue bool
}

// CheckRound -
func (mock *EnableRoundsHandlerMock) CheckRound(round uint64) {
	if mock.CheckRoundCalled != nil {
		mock.CheckRoundCalled(round)
	}
}

// IsCheckValueOnExecByCallerEnabled -
func (mock *EnableRoundsHandlerMock) IsCheckValueOnExecByCallerEnabled() bool {
	return mock.IsCheckValueOnExecByCallerEnabledValue
}

// IsInterfaceNil -
func (mock *EnableRoundsHandlerMock) IsInterfaceNil() bool {
	return mock == nil
}

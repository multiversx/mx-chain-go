package testscommon

// EnableRoundsHandlerStub -
type EnableRoundsHandlerStub struct {
	CheckRoundCalled        func(round uint64)
	IsExampleEnabledChecked func() bool
}

// CheckRound -
func (stub *EnableRoundsHandlerStub) CheckRound(round uint64) {
	if stub.CheckRoundCalled != nil {
		stub.CheckRoundCalled(round)
	}
}

// IsExampleEnabled -
func (stub *EnableRoundsHandlerStub) IsExampleEnabled() bool {
	if stub.IsExampleEnabledChecked != nil {
		return stub.IsExampleEnabledChecked()
	}

	return false
}

// IsInterfaceNil -
func (stub *EnableRoundsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}

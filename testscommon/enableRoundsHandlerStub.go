package testscommon

// EnableRoundsHandlerStub -
type EnableRoundsHandlerStub struct {
	CheckRoundCalled func(round uint64)
}

// CheckRound -
func (stub *EnableRoundsHandlerStub) CheckRound(round uint64) {
	if stub.CheckRoundCalled != nil {
		stub.CheckRoundCalled(round)
	}
}

// IsInterfaceNil -
func (stub *EnableRoundsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}

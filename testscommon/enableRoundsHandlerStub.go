package testscommon

// EnableRoundsHandlerStub -
type EnableRoundsHandlerStub struct {
	CheckRoundCalled                  func(round uint64)
	IsDisableAsyncCallV1EnabledCalled func() bool
}

// CheckRound -
func (stub *EnableRoundsHandlerStub) CheckRound(round uint64) {
	if stub.CheckRoundCalled != nil {
		stub.CheckRoundCalled(round)
	}
}

// IsDisableAsyncCallV1Enabled -
func (stub *EnableRoundsHandlerStub) IsDisableAsyncCallV1Enabled() bool {
	if stub.IsDisableAsyncCallV1EnabledCalled != nil {
		return stub.IsDisableAsyncCallV1EnabledCalled()
	}

	return false
}

// IsInterfaceNil -
func (stub *EnableRoundsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}

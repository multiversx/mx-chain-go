package testscommon

// EnableRoundsHandlerStub -
type EnableRoundsHandlerStub struct {
	IsDisableAsyncCallV1EnabledCalled func() bool
}

// IsDisableAsyncCallV1Enabled -
func (stub *EnableRoundsHandlerStub) IsDisableAsyncCallV1Enabled() bool {
	if stub.IsDisableAsyncCallV1EnabledCalled != nil {
		return stub.IsDisableAsyncCallV1EnabledCalled()
	}

	return false
}

func (stub *EnableRoundsHandlerStub) RoundConfirmed(_ uint64, _ uint64) {
}

// IsInterfaceNil -
func (stub *EnableRoundsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}

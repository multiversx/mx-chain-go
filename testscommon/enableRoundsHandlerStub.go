package testscommon

// EnableRoundsHandlerStub -
type EnableRoundsHandlerStub struct {
	IsDisableAsyncCallV1EnabledCalled func() bool
	RoundConfirmedCalled              func(round uint64, timestamp uint64)
	SupernovaEnableRoundEnabledCalled func() bool
	SupernovaActivationRoundCalled    func() uint64
}

// IsDisableAsyncCallV1Enabled -
func (stub *EnableRoundsHandlerStub) IsDisableAsyncCallV1Enabled() bool {
	if stub.IsDisableAsyncCallV1EnabledCalled != nil {
		return stub.IsDisableAsyncCallV1EnabledCalled()
	}

	return false
}

// RoundConfirmed -
func (stub *EnableRoundsHandlerStub) RoundConfirmed(round uint64, timestamp uint64) {
	if stub.RoundConfirmedCalled != nil {
		stub.RoundConfirmedCalled(round, timestamp)
	}
}

// SupernovaEnableRoundEnabled -
func (stub *EnableRoundsHandlerStub) SupernovaEnableRoundEnabled() bool {
	if stub.SupernovaEnableRoundEnabledCalled != nil {
		return stub.SupernovaEnableRoundEnabledCalled()
	}

	return false
}

// SupernovaActivationRound -
func (stub *EnableRoundsHandlerStub) SupernovaActivationRound() uint64 {
	if stub.SupernovaActivationRoundCalled != nil {
		return stub.SupernovaActivationRoundCalled()
	}

	return 0
}

// IsInterfaceNil -
func (stub *EnableRoundsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}

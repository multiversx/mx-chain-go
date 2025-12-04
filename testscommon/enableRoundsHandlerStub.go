package testscommon

import "github.com/multiversx/mx-chain-go/common"

// EnableRoundsHandlerStub -
type EnableRoundsHandlerStub struct {
	RoundConfirmedCalled       func(round uint64, timestamp uint64)
	IsFlagDefinedCalled        func(flag common.EnableRoundFlag) bool
	IsFlagEnabledCalled        func(flag common.EnableRoundFlag) bool
	IsFlagEnabledInRoundCalled func(flag common.EnableRoundFlag, round uint64) bool
	GetActivationRoundCalled   func(flag common.EnableRoundFlag) uint64
	GetCurrentRoundCalled      func() uint64
}

// RoundConfirmed -
func (e *EnableRoundsHandlerStub) RoundConfirmed(round uint64, timestamp uint64) {
	if e.RoundConfirmedCalled != nil {
		e.RoundConfirmedCalled(round, timestamp)
	}
}

// IsFlagDefined -
func (e *EnableRoundsHandlerStub) IsFlagDefined(flag common.EnableRoundFlag) bool {
	if e.IsFlagDefinedCalled != nil {
		return e.IsFlagDefinedCalled(flag)
	}

	return false
}

// IsFlagEnabled -
func (e *EnableRoundsHandlerStub) IsFlagEnabled(flag common.EnableRoundFlag) bool {
	if e.IsFlagEnabledCalled != nil {
		return e.IsFlagEnabledCalled(flag)
	}

	return false
}

// IsFlagEnabledInRound -
func (e *EnableRoundsHandlerStub) IsFlagEnabledInRound(flag common.EnableRoundFlag, round uint64) bool {
	if e.IsFlagEnabledInRoundCalled != nil {
		return e.IsFlagEnabledInRoundCalled(flag, round)
	}

	return false
}

// GetActivationRound -
func (e *EnableRoundsHandlerStub) GetActivationRound(flag common.EnableRoundFlag) uint64 {
	if e.GetActivationRoundCalled != nil {
		return e.GetActivationRoundCalled(flag)
	}

	return 0
}

// GetCurrentRound -
func (e *EnableRoundsHandlerStub) GetCurrentRound() uint64 {
	if e.GetCurrentRoundCalled != nil {
		return e.GetCurrentRoundCalled()
	}

	return 0
}

// IsInterfaceNil -
func (e *EnableRoundsHandlerStub) IsInterfaceNil() bool {
	return e == nil
}

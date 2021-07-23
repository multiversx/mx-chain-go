package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// EpochStartTriggerStub -
type EpochStartTriggerStub struct {
	ForceEpochStartCalled       func(round uint64) error
	IsEpochStartCalled          func() bool
	EpochCalled                 func() uint32
	ReceivedHeaderCalled        func(handler data.HeaderHandler)
	UpdateCalled                func(round uint64)
	ProcessedCalled             func(header data.HeaderHandler)
	EpochStartRoundCalled       func() uint64
	EpochStartMetaHdrHashCalled func() []byte
}

// SetCurrentEpochStartRound -
func (e *EpochStartTriggerStub) SetCurrentEpochStartRound(_ uint64) {
}

// NotifyAll -
func (e *EpochStartTriggerStub) NotifyAll(_ data.HeaderHandler) {
}

// SetFinalityAttestingRound -
func (e *EpochStartTriggerStub) SetFinalityAttestingRound(_ uint64) {
}

// EpochFinalityAttestingRound -
func (e *EpochStartTriggerStub) EpochFinalityAttestingRound() uint64 {
	return 0
}

// EpochStartMetaHdrHash -
func (e *EpochStartTriggerStub) EpochStartMetaHdrHash() []byte {
	if e.EpochStartMetaHdrHashCalled != nil {
		return e.EpochStartMetaHdrHashCalled()
	}
	return nil
}

// GetRoundsPerEpoch -
func (e *EpochStartTriggerStub) GetRoundsPerEpoch() uint64 {
	return 0
}

// SetTrigger -
func (e *EpochStartTriggerStub) SetTrigger(_ epochStart.TriggerHandler) {
}

// Revert -
func (e *EpochStartTriggerStub) Revert() {
}

// EpochStartRound -
func (e *EpochStartTriggerStub) EpochStartRound() uint64 {
	if e.EpochStartRoundCalled != nil {
		return e.EpochStartRoundCalled()
	}
	return 0
}

// Update -
func (e *EpochStartTriggerStub) Update(round uint64) {
	if e.UpdateCalled != nil {
		e.UpdateCalled(round)
	}
}

// SetProcessed -
func (e *EpochStartTriggerStub) SetProcessed(header data.HeaderHandler) {
	if e.ProcessedCalled != nil {
		e.ProcessedCalled(header)
	}
}

// ForceEpochStart -
func (e *EpochStartTriggerStub) ForceEpochStart(round uint64) error {
	if e.ForceEpochStartCalled != nil {
		return e.ForceEpochStartCalled(round)
	}
	return nil
}

// IsEpochStart -
func (e *EpochStartTriggerStub) IsEpochStart() bool {
	if e.IsEpochStartCalled != nil {
		return e.IsEpochStartCalled()
	}
	return false
}

// Epoch -
func (e *EpochStartTriggerStub) Epoch() uint32 {
	if e.EpochCalled != nil {
		return e.EpochCalled()
	}
	return 0
}

// ReceivedHeader -
func (e *EpochStartTriggerStub) ReceivedHeader(header data.HeaderHandler) {
	if e.ReceivedHeaderCalled != nil {
		e.ReceivedHeaderCalled(header)
	}
}

// SetRoundsPerEpoch -
func (e *EpochStartTriggerStub) SetRoundsPerEpoch(_ uint64) {
}

// SetEpoch -
func (e *EpochStartTriggerStub) SetEpoch(_ uint32) {
}

// IsInterfaceNil -
func (e *EpochStartTriggerStub) IsInterfaceNil() bool {
	return e == nil
}

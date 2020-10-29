package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// EpochStartTriggerStub -
type EpochStartTriggerStub struct {
	ForceEpochStartCalled func()
	IsEpochStartCalled    func() bool
	EpochCalled           func() uint32
	MetaEpochCalled       func() uint32
	ReceivedHeaderCalled  func(handler data.HeaderHandler)
	UpdateCalled          func(round uint64, nonce uint64)
	ProcessedCalled       func(header data.HeaderHandler)
	EpochStartRoundCalled func() uint64
}

// RevertStateToBlock -
func (e *EpochStartTriggerStub) RevertStateToBlock(_ data.HeaderHandler) error {
	return nil
}

// RequestEpochStartIfNeeded -
func (e *EpochStartTriggerStub) RequestEpochStartIfNeeded(_ data.HeaderHandler) {
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
	return nil
}

// GetSavedStateKey -
func (e *EpochStartTriggerStub) GetSavedStateKey() []byte {
	return []byte("epoch start trigger key")
}

// LoadState -
func (e *EpochStartTriggerStub) LoadState(_ []byte) error {
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
func (e *EpochStartTriggerStub) Revert(_ data.HeaderHandler) {
}

// SetAppStatusHandler -
func (e *EpochStartTriggerStub) SetAppStatusHandler(_ core.AppStatusHandler) error {
	return nil
}

// EpochStartRound -
func (e *EpochStartTriggerStub) EpochStartRound() uint64 {
	if e.EpochStartRoundCalled != nil {
		return e.EpochStartRoundCalled()
	}
	return 0
}

// Update -
func (e *EpochStartTriggerStub) Update(round uint64, nonce uint64) {
	if e.UpdateCalled != nil {
		e.UpdateCalled(round, nonce)
	}
}

// SetProcessed -
func (e *EpochStartTriggerStub) SetProcessed(header data.HeaderHandler, _ data.BodyHandler) {
	if e.ProcessedCalled != nil {
		e.ProcessedCalled(header)
	}
}

// ForceEpochStart -
func (e *EpochStartTriggerStub) ForceEpochStart() {
	if e.ForceEpochStartCalled != nil {
		e.ForceEpochStartCalled()
	}
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

// MetaEpoch -
func (e *EpochStartTriggerStub) MetaEpoch() uint32 {
	if e.MetaEpochCalled != nil {
		return e.MetaEpochCalled()
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

// Close -
func (e *EpochStartTriggerStub) Close() error {
	return nil
}

// SetEpoch -
func (e *EpochStartTriggerStub) SetEpoch(_ uint32) {
}

// IsInterfaceNil -
func (e *EpochStartTriggerStub) IsInterfaceNil() bool {
	return e == nil
}

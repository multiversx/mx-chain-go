package mock

import "github.com/ElrondNetwork/elrond-go/data"

// EpochStartTriggerStub -
type EpochStartTriggerStub struct {
	ForceEpochStartCalled func(round uint64) error
	IsEpochStartCalled    func() bool
	EpochCalled           func() uint32
	ReceivedHeaderCalled  func(handler data.HeaderHandler)
	UpdateCalled          func(round uint64)
	ProcessedCalled       func(header data.HeaderHandler)
	EpochStartRoundCalled func() uint64
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

// GetSavedStateKey -
func (e *EpochStartTriggerStub) GetSavedStateKey() []byte {
	return []byte("key")
}

// LoadState -
func (e *EpochStartTriggerStub) LoadState(_ []byte) error {
	return nil
}

// EpochStartMetaHdrHash -
func (e *EpochStartTriggerStub) EpochStartMetaHdrHash() []byte {
	return nil
}

// Revert -
func (e *EpochStartTriggerStub) Revert(_ uint64) {
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

// IsInterfaceNil -
func (e *EpochStartTriggerStub) IsInterfaceNil() bool {
	return e == nil
}

package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

type epochStartTrigger struct {
}

// NewEpochStartTrigger returns a new instance of epochStartTrigger
func NewEpochStartTrigger() *epochStartTrigger {
	return &epochStartTrigger{}
}

// Update -
func (e *epochStartTrigger) Update(_ uint64, _ uint64) {
}

// ShouldProposeEpochChange -
func (e *epochStartTrigger) ShouldProposeEpochChange(_ uint64, _ uint64) bool {
	return false
}

// SetEpochChange -
func (e *epochStartTrigger) SetEpochChange(_ uint64) {}

// ReceivedHeader -
func (e *epochStartTrigger) ReceivedHeader(_ data.HeaderHandler) {
}

// IsEpochStart -
func (e *epochStartTrigger) IsEpochStart() bool {
	return false
}

// Epoch -
func (e *epochStartTrigger) Epoch() uint32 {
	return 0
}

// MetaEpoch -
func (e *epochStartTrigger) MetaEpoch() uint32 {
	return 0
}

// EpochStartRound -
func (e *epochStartTrigger) EpochStartRound() uint64 {
	return 0
}

// SetProcessed -
func (e *epochStartTrigger) SetProcessed(_ data.HeaderHandler, _ data.BodyHandler) {
}

// RevertStateToBlock -
func (e *epochStartTrigger) RevertStateToBlock(_ data.HeaderHandler) error {
	return nil
}

// EpochStartMetaHdrHash -
func (e *epochStartTrigger) EpochStartMetaHdrHash() []byte {
	return nil
}

// LastCommitedEpochStartHdr -
func (e *epochStartTrigger) LastCommitedEpochStartHdr() (data.HeaderHandler, error) {
	return &block.HeaderV2{}, nil
}

// GetEpochStartHdrFromStorage -
func (e *epochStartTrigger) GetEpochStartHdrFromStorage(_ uint32) (data.HeaderHandler, error) {
	return &block.HeaderV2{}, nil
}

// GetSavedStateKey -
func (e *epochStartTrigger) GetSavedStateKey() []byte {
	return nil
}

// LoadState -
func (e *epochStartTrigger) LoadState(_ []byte) error {
	return nil
}

// SetFinalityAttestingRound -
func (e *epochStartTrigger) SetFinalityAttestingRound(_ uint64) {
}

// EpochFinalityAttestingRound -
func (e *epochStartTrigger) EpochFinalityAttestingRound() uint64 {
	return 0
}

// RequestEpochStartIfNeeded -
func (e *epochStartTrigger) RequestEpochStartIfNeeded(_ data.HeaderHandler) {
}

// IsInterfaceNil -
func (e *epochStartTrigger) IsInterfaceNil() bool {
	return e == nil
}

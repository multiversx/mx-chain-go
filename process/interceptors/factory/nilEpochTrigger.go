package factory

import "github.com/ElrondNetwork/elrond-go/data"

// nil implementation for epoch trigger
type nilEpochTrigger struct {
}

// RequestEpochStartIfNeeded -
func (n *nilEpochTrigger) RequestEpochStartIfNeeded(_ data.HeaderHandler) {
}

// Update -
func (n *nilEpochTrigger) Update(_ uint64) {
}

// ReceivedHeader -
func (n *nilEpochTrigger) ReceivedHeader(_ data.HeaderHandler) {
}

// IsEpochStart -
func (n *nilEpochTrigger) IsEpochStart() bool {
	return false
}

// Epoch -
func (n *nilEpochTrigger) Epoch() uint32 {
	return 0
}

// EpochStartRound -
func (n *nilEpochTrigger) EpochStartRound() uint64 {
	return 0
}

// SetProcessed -
func (n *nilEpochTrigger) SetProcessed(_ data.HeaderHandler) {
}

// Revert -
func (n *nilEpochTrigger) Revert(_ uint64) {
}

// EpochStartMetaHdrHash -
func (n *nilEpochTrigger) EpochStartMetaHdrHash() []byte {
	return nil
}

// IsInterfaceNil -
func (n *nilEpochTrigger) IsInterfaceNil() bool {
	return n == nil
}

// SetFinalityAttestingRound -
func (n *nilEpochTrigger) SetFinalityAttestingRound(_ uint64) {
}

// EpochFinalityAttestingRound -
func (n *nilEpochTrigger) EpochFinalityAttestingRound() uint64 {
	return 0
}

func (n *nilEpochTrigger) GetSavedStateKey() []byte {
	return nil
}

func (n *nilEpochTrigger) LoadState(_ []byte) error {
	return nil
}

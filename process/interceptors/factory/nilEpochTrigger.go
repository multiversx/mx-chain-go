package factory

import "github.com/ElrondNetwork/elrond-go/data"

// nil implementation for epoch trigger
type nilEpochTrigger struct {
}

func (n *nilEpochTrigger) Update(_ uint64) {
}

func (n *nilEpochTrigger) ReceivedHeader(_ data.HeaderHandler) {
}

func (n *nilEpochTrigger) IsEpochStart() bool {
	return false
}

func (n *nilEpochTrigger) Epoch() uint32 {
	return 0
}

func (n *nilEpochTrigger) EpochStartRound() uint64 {
	return 0
}

func (n *nilEpochTrigger) SetProcessed(_ data.HeaderHandler) {
}

func (n *nilEpochTrigger) Revert(_ uint64) {
}

func (n *nilEpochTrigger) EpochStartMetaHdrHash() []byte {
	return nil
}

func (n *nilEpochTrigger) IsInterfaceNil() bool {
	return n == nil
}

func (n *nilEpochTrigger) SetFinalityAttestingRound(_ uint64) {
}

func (n *nilEpochTrigger) EpochFinalityAttestingRound() uint64 {
	return 0
}

func (n *nilEpochTrigger) GetSavedStateKey() []byte {
	return nil
}

func (n *nilEpochTrigger) LoadState(_ []byte) error {
	return nil
}

package disabled

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type epochStartTrigger struct {
}

// NewEpochStartTrigger returns a new instance of epochStartTrigger
func NewEpochStartTrigger() *epochStartTrigger {
	return &epochStartTrigger{}
}

func (e *epochStartTrigger) Update(round uint64) {
}

func (e *epochStartTrigger) ReceivedHeader(header data.HeaderHandler) {
}

func (e *epochStartTrigger) IsEpochStart() bool {
	return false
}

func (e *epochStartTrigger) Epoch() uint32 {
	return 0
}

func (e *epochStartTrigger) EpochStartRound() uint64 {
	return 0
}

func (e *epochStartTrigger) SetProcessed(header data.HeaderHandler) {
}

func (e *epochStartTrigger) RevertStateToBlock(header data.HeaderHandler) error {
	return nil
}

func (e *epochStartTrigger) EpochStartMetaHdrHash() []byte {
	return nil
}

func (e *epochStartTrigger) GetSavedStateKey() []byte {
	return nil
}

func (e *epochStartTrigger) LoadState(key []byte) error {
	return nil
}

func (e *epochStartTrigger) SetFinalityAttestingRound(round uint64) {
}

func (e *epochStartTrigger) EpochFinalityAttestingRound() uint64 {
	return 0
}

func (e *epochStartTrigger) RequestEpochStartIfNeeded(interceptedHeader data.HeaderHandler) {
}

func (e *epochStartTrigger) IsInterfaceNil() bool {
	return e == nil
}

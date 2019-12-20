package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

type EpochStartTriggerStub struct {
	ForceEpochStartCalled func(round uint64) error
	IsEpochStartCalled    func() bool
	EpochCalled           func() uint32
	ReceivedHeaderCalled  func(handler data.HeaderHandler)
	UpdateCalled          func(round uint64)
	ProcessedCalled       func(header data.HeaderHandler)
	EpochStartRoundCalled func() uint64
}

func (e *EpochStartTriggerStub) SetFinalityAttestingRound(round uint64) {
}

func (e *EpochStartTriggerStub) EpochFinalityAttestingRound() uint64 {
	return 0
}

func (e *EpochStartTriggerStub) EpochStartMetaHdrHash() []byte {
	return nil
}

func (e *EpochStartTriggerStub) GetRoundsPerEpoch() uint64 {
	return 0
}

func (e *EpochStartTriggerStub) SetTrigger(triggerHandler epochStart.TriggerHandler) {
}

func (e *EpochStartTriggerStub) Revert() {
}

func (e *EpochStartTriggerStub) EpochStartRound() uint64 {
	if e.EpochStartRoundCalled != nil {
		return e.EpochStartRoundCalled()
	}
	return 0
}

func (e *EpochStartTriggerStub) Update(round uint64) {
	if e.UpdateCalled != nil {
		e.UpdateCalled(round)
	}
}

func (e *EpochStartTriggerStub) SetProcessed(header data.HeaderHandler) {
	if e.ProcessedCalled != nil {
		e.ProcessedCalled(header)
	}
}

func (e *EpochStartTriggerStub) ForceEpochStart(round uint64) error {
	if e.ForceEpochStartCalled != nil {
		return e.ForceEpochStartCalled(round)
	}
	return nil
}

func (e *EpochStartTriggerStub) IsEpochStart() bool {
	if e.IsEpochStartCalled != nil {
		return e.IsEpochStartCalled()
	}
	return false
}

func (e *EpochStartTriggerStub) Epoch() uint32 {
	if e.EpochCalled != nil {
		return e.EpochCalled()
	}
	return 0
}

func (e *EpochStartTriggerStub) ReceivedHeader(header data.HeaderHandler) {
	if e.ReceivedHeaderCalled != nil {
		e.ReceivedHeaderCalled(header)
	}
}

func (e *EpochStartTriggerStub) SetRoundsPerEpoch(roundsPerEpoch uint64) {
}

func (e *EpochStartTriggerStub) IsInterfaceNil() bool {
	return e == nil
}

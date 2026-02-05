package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/process"
)

// AOTSelectorStub -
type AOTSelectorStub struct {
	TriggerAOTSelectionCalled        func(committedHeader data.HeaderHandler, currentRound uint64)
	GetPreSelectedTransactionsCalled func(blockNonce uint64) (*process.AOTSelectionResult, bool)
	CancelOngoingSelectionCalled     func()
}

// TriggerAOTSelection -
func (stub *AOTSelectorStub) TriggerAOTSelection(committedHeader data.HeaderHandler, currentRound uint64) {
	if stub.TriggerAOTSelectionCalled != nil {
		stub.TriggerAOTSelectionCalled(committedHeader, currentRound)
	}
}

// GetPreSelectedTransactions -
func (stub *AOTSelectorStub) GetPreSelectedTransactions(blockNonce uint64) (*process.AOTSelectionResult, bool) {
	if stub.GetPreSelectedTransactionsCalled != nil {
		return stub.GetPreSelectedTransactionsCalled(blockNonce)
	}
	return nil, false
}

// CancelOngoingSelection -
func (stub *AOTSelectorStub) CancelOngoingSelection() {
	if stub.CancelOngoingSelectionCalled != nil {
		stub.CancelOngoingSelectionCalled()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *AOTSelectorStub) IsInterfaceNil() bool {
	return stub == nil
}

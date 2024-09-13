package state

import (
	"github.com/multiversx/mx-chain-core-go/data/stateChange"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/state/stateChanges"
)

// StateChangesCollectorStub represents a mock for the StateChangesCollector interface
type StateChangesCollectorStub struct {
	AddStateChangeCalled                   func(stateChange stateChanges.StateChange)
	AddSaveAccountStateChangeCalled        func(oldAccount, account vmcommon.AccountHandler, stateChange stateChanges.StateChange)
	ResetCalled                            func()
	AddTxHashToCollectedStateChangesCalled func(txHash []byte, tx *transaction.Transaction)
	SetIndexToLastStateChangeCalled        func(index int) error
	RevertToIndexCalled                    func(index int) error
	PublishCalled                          func() error
	IsInterfaceNilCalled                   func() bool
	GetStateChangesForTxsCalled            func() map[string]*stateChange.StateChanges
}

// AddStateChange -
func (s *StateChangesCollectorStub) AddStateChange(stateChange stateChanges.StateChange) {
	if s.AddStateChangeCalled != nil {
		s.AddStateChangeCalled(stateChange)
	}
}

// AddSaveAccountStateChange -
func (s *StateChangesCollectorStub) AddSaveAccountStateChange(oldAccount, account vmcommon.AccountHandler, stateChange stateChanges.StateChange) {
	if s.AddSaveAccountStateChangeCalled != nil {
		s.AddSaveAccountStateChangeCalled(oldAccount, account, stateChange)
	}
}

// Reset -
func (s *StateChangesCollectorStub) Reset() {
	if s.ResetCalled != nil {
		s.ResetCalled()
	}
}

// AddTxHashToCollectedStateChanges -
func (s *StateChangesCollectorStub) AddTxHashToCollectedStateChanges(txHash []byte, tx *transaction.Transaction) {
	if s.AddTxHashToCollectedStateChangesCalled != nil {
		s.AddTxHashToCollectedStateChangesCalled(txHash, tx)
	}
}

// SetIndexToLastStateChange -
func (s *StateChangesCollectorStub) SetIndexToLastStateChange(index int) error {
	if s.SetIndexToLastStateChangeCalled != nil {
		return s.SetIndexToLastStateChangeCalled(index)
	}

	return nil
}

// RevertToIndex -
func (s *StateChangesCollectorStub) RevertToIndex(index int) error {
	if s.RevertToIndexCalled != nil {
		return s.RevertToIndexCalled(index)
	}

	return nil
}

// Publish -
func (s *StateChangesCollectorStub) Publish() error {
	if s.PublishCalled != nil {
		return s.PublishCalled()
	}

	return nil
}

// IsInterfaceNil -
func (s *StateChangesCollectorStub) IsInterfaceNil() bool {
	if s.IsInterfaceNilCalled != nil {
		return s.IsInterfaceNilCalled()
	}

	return false
}

// GetStateChangesForTxs -
func (s *StateChangesCollectorStub) GetStateChangesForTxs() map[string]*stateChange.StateChanges {
	if s.GetStateChangesForTxsCalled != nil {
		return s.GetStateChangesForTxsCalled()
	}

	return nil
}

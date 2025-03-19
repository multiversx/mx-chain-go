package state

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/stateChange"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/state"
)

// StateChangesCollectorStub represents a mock for the StateChangesCollector interface
type StateChangesCollectorStub struct {
	AddStateChangeCalled                   func(stateChange state.StateChange)
	AddSaveAccountStateChangeCalled        func(oldAccount, account vmcommon.AccountHandler, stateChange state.StateChange)
	ResetCalled                            func()
	AddTxHashToCollectedStateChangesCalled func(txHash []byte, tx data.TransactionHandler)
	SetIndexToLastStateChangeCalled        func(index int) error
	RevertToIndexCalled                    func(index int) error
	PublishCalled                          func() (map[string]*stateChange.StateChanges, error)
	StoreCalled                            func() error
	IsInterfaceNilCalled                   func() bool
}

// AddStateChange -
func (s *StateChangesCollectorStub) AddStateChange(stateChange state.StateChange) {
	if s.AddStateChangeCalled != nil {
		s.AddStateChangeCalled(stateChange)
	}
}

// AddSaveAccountStateChange -
func (s *StateChangesCollectorStub) AddSaveAccountStateChange(oldAccount, account vmcommon.AccountHandler, stateChange state.StateChange) {
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
func (s *StateChangesCollectorStub) AddTxHashToCollectedStateChanges(txHash []byte, tx data.TransactionHandler) {
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
func (s *StateChangesCollectorStub) Publish() (map[string]*stateChange.StateChanges, error) {
	if s.PublishCalled != nil {
		return s.PublishCalled()
	}

	return nil, nil
}

func (s *StateChangesCollectorStub) Store() error {
	if s.StoreCalled != nil {
		return s.StoreCalled()
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

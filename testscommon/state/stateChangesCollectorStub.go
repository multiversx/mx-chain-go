package state

import (
	"github.com/multiversx/mx-chain-core-go/data/stateChange"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// StateAccessesCollectorStub represents a mock for the StateAccessesCollector interface
type StateAccessesCollectorStub struct {
	AddStateChangeCalled                   func(stateAccess *stateChange.StateAccess)
	AddSaveAccountStateChangeCalled        func(oldAccount, account vmcommon.AccountHandler, stateChange *stateChange.StateAccess)
	ResetCalled                            func()
	AddTxHashToCollectedStateChangesCalled func(txHash []byte)
	SetIndexToLastStateChangeCalled        func(index int) error
	RevertToIndexCalled                    func(index int) error
	PublishCalled                          func() (map[string]*stateChange.StateAccesses, error)
	StoreCalled                            func() error
	IsInterfaceNilCalled                   func() bool
}

// AddStateAccess -
func (s *StateAccessesCollectorStub) AddStateAccess(stateChange *stateChange.StateAccess) {
	if s.AddStateChangeCalled != nil {
		s.AddStateChangeCalled(stateChange)
	}
}

// AddSaveAccountStateAccess -
func (s *StateAccessesCollectorStub) AddSaveAccountStateAccess(oldAccount, account vmcommon.AccountHandler, stateChange *stateChange.StateAccess) {
	if s.AddSaveAccountStateChangeCalled != nil {
		s.AddSaveAccountStateChangeCalled(oldAccount, account, stateChange)
	}
}

// Reset -
func (s *StateAccessesCollectorStub) Reset() {
	if s.ResetCalled != nil {
		s.ResetCalled()
	}
}

// AddTxHashToCollectedStateChanges -
func (s *StateAccessesCollectorStub) AddTxHashToCollectedStateChanges(txHash []byte) {
	if s.AddTxHashToCollectedStateChangesCalled != nil {
		s.AddTxHashToCollectedStateChangesCalled(txHash)
	}
}

// SetIndexToLastStateChange -
func (s *StateAccessesCollectorStub) SetIndexToLastStateChange(index int) error {
	if s.SetIndexToLastStateChangeCalled != nil {
		return s.SetIndexToLastStateChangeCalled(index)
	}

	return nil
}

// RevertToIndex -
func (s *StateAccessesCollectorStub) RevertToIndex(index int) error {
	if s.RevertToIndexCalled != nil {
		return s.RevertToIndexCalled(index)
	}

	return nil
}

// Publish -
func (s *StateAccessesCollectorStub) Publish() (map[string]*stateChange.StateAccesses, error) {
	if s.PublishCalled != nil {
		return s.PublishCalled()
	}

	return nil, nil
}

func (s *StateAccessesCollectorStub) Store() error {
	if s.StoreCalled != nil {
		return s.StoreCalled()
	}

	return nil
}

// IsInterfaceNil -
func (s *StateAccessesCollectorStub) IsInterfaceNil() bool {
	if s.IsInterfaceNilCalled != nil {
		return s.IsInterfaceNilCalled()
	}

	return false
}

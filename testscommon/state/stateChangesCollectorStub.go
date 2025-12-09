package state

import (
	"github.com/multiversx/mx-chain-core-go/data/stateChange"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// StateAccessesCollectorStub represents a mock for the StateAccessesCollector interface
type StateAccessesCollectorStub struct {
	AddStateChangeCalled                    func(stateAccess *stateChange.StateAccess)
	GetAccountChangesCalled                 func(oldAccount, account vmcommon.AccountHandler) uint32
	ResetCalled                             func()
	AddTxHashToCollectedStateAccessesCalled func(txHash []byte)
	SetIndexToLatestStateAccessesCalled     func(index int) error
	RevertToIndexCalled                     func(index int) error
	GetCollectedAccessesCalled              func() map[string]*stateChange.StateAccesses
	StoreCalled                             func() error
	IsInterfaceNilCalled                    func() bool
}

// AddStateAccess -
func (s *StateAccessesCollectorStub) AddStateAccess(stateChange *stateChange.StateAccess) {
	if s.AddStateChangeCalled != nil {
		s.AddStateChangeCalled(stateChange)
	}
}

// GetAccountChanges -
func (s *StateAccessesCollectorStub) GetAccountChanges(oldAccount, account vmcommon.AccountHandler) uint32 {
	if s.GetAccountChangesCalled != nil {
		return s.GetAccountChangesCalled(oldAccount, account)
	}
	return stateChange.NoChange
}

// Reset -
func (s *StateAccessesCollectorStub) Reset() {
	if s.ResetCalled != nil {
		s.ResetCalled()
	}
}

// AddTxHashToCollectedStateAccesses -
func (s *StateAccessesCollectorStub) AddTxHashToCollectedStateAccesses(txHash []byte) {
	if s.AddTxHashToCollectedStateAccessesCalled != nil {
		s.AddTxHashToCollectedStateAccessesCalled(txHash)
	}
}

// SetIndexToLatestStateAccesses -
func (s *StateAccessesCollectorStub) SetIndexToLatestStateAccesses(index int) error {
	if s.SetIndexToLatestStateAccessesCalled != nil {
		return s.SetIndexToLatestStateAccessesCalled(index)
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

// GetCollectedAccesses -
func (s *StateAccessesCollectorStub) GetCollectedAccesses() map[string]*stateChange.StateAccesses {
	if s.GetCollectedAccessesCalled != nil {
		return s.GetCollectedAccessesCalled()
	}

	return nil
}

// Store -
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

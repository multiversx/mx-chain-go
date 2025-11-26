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
	GetStateAccessesForRootHashCalled       func(rootHash []byte) map[string]*stateChange.StateAccesses
	RemoveStateAccessesForRootHashCalled    func(rootHash []byte)
	CommitCollectedAccessesCalled           func(rootHash []byte) error
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

// GetStateAccessesForRootHash -
func (s *StateAccessesCollectorStub) GetStateAccessesForRootHash(rootHash []byte) map[string]*stateChange.StateAccesses {
	if s.GetStateAccessesForRootHashCalled != nil {
		return s.GetStateAccessesForRootHashCalled(rootHash)
	}

	return nil
}

// RemoveStateAccessesForRootHash -
func (s *StateAccessesCollectorStub) RemoveStateAccessesForRootHash(rootHash []byte) {
	if s.RemoveStateAccessesForRootHashCalled != nil {
		s.RemoveStateAccessesForRootHashCalled(rootHash)
	}
}

// CommitCollectedAccesses -
func (s *StateAccessesCollectorStub) CommitCollectedAccesses(rootHash []byte) error {
	if s.CommitCollectedAccessesCalled != nil {
		return s.CommitCollectedAccessesCalled(rootHash)
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

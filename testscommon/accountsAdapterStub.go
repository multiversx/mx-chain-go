package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/mock"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// AccountsStub -
type AccountsStub struct {
	GetExistingAccountCalled func(addressContainer []byte) (vmcommon.AccountHandler, error)
	LoadAccountCalled        func(container []byte) (vmcommon.AccountHandler, error)
	SaveAccountCalled        func(account vmcommon.AccountHandler) error
	RemoveAccountCalled      func(addressContainer []byte) error
	CommitCalled             func() ([]byte, error)
	JournalLenCalled         func() int
	RevertToSnapshotCalled   func(snapshot int) error
	RootHashCalled           func() ([]byte, error)
	RecreateTrieCalled       func(rootHash []byte) error
	PruneTrieCalled          func(rootHash []byte, identifier temporary.TriePruningIdentifier)
	CancelPruneCalled        func(rootHash []byte, identifier temporary.TriePruningIdentifier)
	SnapshotStateCalled      func(rootHash []byte)
	SetStateCheckpointCalled func(rootHash []byte)
	IsPruningEnabledCalled   func() bool
	GetAllLeavesCalled       func(rootHash []byte) (chan core.KeyValueHolder, error)
	RecreateAllTriesCalled   func(rootHash []byte) (map[string]temporary.Trie, error)
	GetNumCheckpointsCalled  func() uint32
	GetCodeCalled            func([]byte) []byte
	GetTrieCalled            func([]byte) (temporary.Trie, error)
}

// GetTrie -
func (as *AccountsStub) GetTrie(codeHash []byte) (temporary.Trie, error) {
	if as.GetTrieCalled != nil {
		return as.GetTrieCalled(codeHash)
	}

	return nil, nil
}

// GetCode -
func (as *AccountsStub) GetCode(codeHash []byte) []byte {
	if as.GetCodeCalled != nil {
		return as.GetCodeCalled(codeHash)
	}
	return nil
}

// RecreateAllTries -
func (as *AccountsStub) RecreateAllTries(rootHash []byte) (map[string]temporary.Trie, error) {
	if as.RecreateAllTriesCalled != nil {
		return as.RecreateAllTriesCalled(rootHash)
	}
	return nil, nil
}

// LoadAccount -
func (as *AccountsStub) LoadAccount(address []byte) (vmcommon.AccountHandler, error) {
	if as.LoadAccountCalled != nil {
		return as.LoadAccountCalled(address)
	}
	return mock.NewAccountWrapMock(address), nil
}

// SaveAccount -
func (as *AccountsStub) SaveAccount(account vmcommon.AccountHandler) error {
	if as.SaveAccountCalled != nil {
		return as.SaveAccountCalled(account)
	}
	return nil
}

// GetAllLeaves -
func (as *AccountsStub) GetAllLeaves(rootHash []byte) (chan core.KeyValueHolder, error) {
	if as.GetAllLeavesCalled != nil {
		return as.GetAllLeavesCalled(rootHash)
	}
	return nil, nil
}

// Commit -
func (as *AccountsStub) Commit() ([]byte, error) {
	if as.CommitCalled != nil {
		return as.CommitCalled()
	}

	return nil, errNotImplemented
}

// GetExistingAccount -
func (as *AccountsStub) GetExistingAccount(addressContainer []byte) (vmcommon.AccountHandler, error) {
	if as.GetExistingAccountCalled != nil {
		return as.GetExistingAccountCalled(addressContainer)
	}

	return nil, errNotImplemented
}

// JournalLen -
func (as *AccountsStub) JournalLen() int {
	if as.JournalLenCalled != nil {
		return as.JournalLenCalled()
	}

	return 0
}

// RemoveAccount -
func (as *AccountsStub) RemoveAccount(addressContainer []byte) error {
	if as.RemoveAccountCalled != nil {
		return as.RemoveAccountCalled(addressContainer)
	}

	return errNotImplemented
}

// RevertToSnapshot -
func (as *AccountsStub) RevertToSnapshot(snapshot int) error {
	if as.RevertToSnapshotCalled != nil {
		return as.RevertToSnapshotCalled(snapshot)
	}

	return errNotImplemented
}

// RootHash -
func (as *AccountsStub) RootHash() ([]byte, error) {
	if as.RootHashCalled != nil {
		return as.RootHashCalled()
	}

	return nil, errNotImplemented
}

// RecreateTrie -
func (as *AccountsStub) RecreateTrie(rootHash []byte) error {
	if as.RecreateTrieCalled != nil {
		return as.RecreateTrieCalled(rootHash)
	}

	return errNotImplemented
}

// PruneTrie -
func (as *AccountsStub) PruneTrie(rootHash []byte, identifier temporary.TriePruningIdentifier) {
	as.PruneTrieCalled(rootHash, identifier)
}

// CancelPrune -
func (as *AccountsStub) CancelPrune(rootHash []byte, identifier temporary.TriePruningIdentifier) {
	if as.CancelPruneCalled != nil {
		as.CancelPruneCalled(rootHash, identifier)
	}
}

// SnapshotState -
func (as *AccountsStub) SnapshotState(rootHash []byte) {
	if as.SnapshotStateCalled != nil {
		as.SnapshotStateCalled(rootHash)
	}
}

// SetStateCheckpoint -
func (as *AccountsStub) SetStateCheckpoint(rootHash []byte) {
	if as.SetStateCheckpointCalled != nil {
		as.SetStateCheckpointCalled(rootHash)
	}
}

// IsPruningEnabled -
func (as *AccountsStub) IsPruningEnabled() bool {
	if as.IsPruningEnabledCalled != nil {
		return as.IsPruningEnabledCalled()
	}

	return false
}

// GetNumCheckpoints -
func (as *AccountsStub) GetNumCheckpoints() uint32 {
	if as.GetNumCheckpointsCalled != nil {
		return as.GetNumCheckpointsCalled()
	}

	return 0
}

// Close -
func (as *AccountsStub) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (as *AccountsStub) IsInterfaceNil() bool {
	return as == nil
}

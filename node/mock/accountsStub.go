package mock

import (
	"context"
	"errors"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// AccountsStub -
type AccountsStub struct {
	GetExistingAccountCalled func(addressContainer []byte) (state.AccountHandler, error)
	LoadAccountCalled        func(container []byte) (state.AccountHandler, error)
	SaveAccountCalled        func(account state.AccountHandler) error
	RemoveAccountCalled      func(addressContainer []byte) error
	CommitCalled             func() ([]byte, error)
	JournalLenCalled         func() int
	RevertToSnapshotCalled   func(snapshot int) error
	RootHashCalled           func() ([]byte, error)
	RecreateTrieCalled       func(rootHash []byte) error
	PruneTrieCalled          func(rootHash []byte, identifier data.TriePruningIdentifier)
	CancelPruneCalled        func(rootHash []byte, identifier data.TriePruningIdentifier)
	SnapshotStateCalled      func(rootHash []byte)
	SetStateCheckpointCalled func(rootHash []byte)
	IsPruningEnabledCalled   func() bool
	GetAllLeavesCalled       func(rootHash []byte) (chan core.KeyValueHolder, error)
	RecreateAllTriesCalled   func(rootHash []byte) (map[string]data.Trie, error)
	GetNumCheckpointsCalled  func() uint32
	GetCodeCalled            func(state.AccountHandler) []byte
}

// GetCode -
func (as *AccountsStub) GetCode(account state.AccountHandler) []byte {
	if as.GetCodeCalled != nil {
		return as.GetCodeCalled(account)
	}
	return nil
}

// RecreateAllTries -
func (as *AccountsStub) RecreateAllTries(rootHash []byte, _ context.Context) (map[string]data.Trie, error) {
	if as.RecreateAllTriesCalled != nil {
		return as.RecreateAllTriesCalled(rootHash)
	}
	return nil, nil
}

// LoadAccount -
func (as *AccountsStub) LoadAccount(address []byte) (state.AccountHandler, error) {
	if as.LoadAccountCalled != nil {
		return as.LoadAccountCalled(address)
	}
	return nil, nil
}

// SaveAccount -
func (as *AccountsStub) SaveAccount(account state.AccountHandler) error {
	if as.SaveAccountCalled != nil {
		return as.SaveAccountCalled(account)
	}
	return nil
}

// GetAllLeaves -
func (as *AccountsStub) GetAllLeaves(rootHash []byte, _ context.Context) (chan core.KeyValueHolder, error) {
	if as.GetAllLeavesCalled != nil {
		return as.GetAllLeavesCalled(rootHash)
	}
	return nil, nil
}

var errNotImplemented = errors.New("not implemented")

// Commit -
func (as *AccountsStub) Commit() ([]byte, error) {
	if as.CommitCalled != nil {
		return as.CommitCalled()
	}

	return nil, errNotImplemented
}

// GetExistingAccount -
func (as *AccountsStub) GetExistingAccount(addressContainer []byte) (state.AccountHandler, error) {
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
func (as *AccountsStub) PruneTrie(rootHash []byte, identifier data.TriePruningIdentifier) {
	as.PruneTrieCalled(rootHash, identifier)
}

// CancelPrune -
func (as *AccountsStub) CancelPrune(rootHash []byte, identifier data.TriePruningIdentifier) {
	if as.CancelPruneCalled != nil {
		as.CancelPruneCalled(rootHash, identifier)
	}
}

// SnapshotState -
func (as *AccountsStub) SnapshotState(rootHash []byte, _ context.Context) {
	if as.SnapshotStateCalled != nil {
		as.SnapshotStateCalled(rootHash)
	}
}

// SetStateCheckpoint -
func (as *AccountsStub) SetStateCheckpoint(rootHash []byte, _ context.Context) {
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

// IsInterfaceNil returns true if there is no value under the interface
func (as *AccountsStub) IsInterfaceNil() bool {
	return as == nil
}

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
	GetExistingAccountCalled func(address []byte) (state.AccountHandler, error)
	LoadAccountCalled        func(address []byte) (state.AccountHandler, error)
	SaveAccountCalled        func(account state.AccountHandler) error
	RemoveAccountCalled      func(address []byte) error
	CommitCalled             func() ([]byte, error)
	JournalLenCalled         func() int
	RevertToSnapshotCalled   func(snapshot int) error
	RootHashCalled           func() ([]byte, error)
	RecreateTrieCalled       func(rootHash []byte) error
	PruneTrieCalled          func(rootHash []byte, identifier data.TriePruningIdentifier)
	CancelPruneCalled        func(rootHash []byte, identifier data.TriePruningIdentifier)
	SnapshotStateCalled      func(rootHash []byte, ctx context.Context)
	SetStateCheckpointCalled func(rootHash []byte, ctx context.Context)
	IsPruningEnabledCalled   func() bool
	GetAllLeavesCalled       func(rootHash []byte, ctx context.Context) (chan core.KeyValueHolder, error)
}

// GetNumCheckpoints -
func (as *AccountsStub) GetNumCheckpoints() uint32 {
	panic("implement me")
}

// RecreateAllTries -
func (as *AccountsStub) RecreateAllTries(_ []byte, _ context.Context) (map[string]data.Trie, error) {
	panic("implement me")
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
func (as *AccountsStub) GetAllLeaves(rootHash []byte, ctx context.Context) (chan core.KeyValueHolder, error) {
	if as.GetAllLeavesCalled != nil {
		return as.GetAllLeavesCalled(rootHash, ctx)
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
func (as *AccountsStub) GetExistingAccount(address []byte) (state.AccountHandler, error) {
	if as.GetExistingAccountCalled != nil {
		return as.GetExistingAccountCalled(address)
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
func (as *AccountsStub) RemoveAccount(address []byte) error {
	if as.RemoveAccountCalled != nil {
		return as.RemoveAccountCalled(address)
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
	if as.PruneTrieCalled != nil {
		as.PruneTrieCalled(rootHash, identifier)
	}
}

// CancelPrune -
func (as *AccountsStub) CancelPrune(rootHash []byte, identifier data.TriePruningIdentifier) {
	if as.CancelPruneCalled != nil {
		as.CancelPruneCalled(rootHash, identifier)
	}
}

// SnapshotState -
func (as *AccountsStub) SnapshotState(rootHash []byte, ctx context.Context) {
	if as.SnapshotStateCalled != nil {
		as.SnapshotStateCalled(rootHash, ctx)
	}
}

// SetStateCheckpoint -
func (as *AccountsStub) SetStateCheckpoint(rootHash []byte, ctx context.Context) {
	if as.SetStateCheckpointCalled != nil {
		as.SetStateCheckpointCalled(rootHash, ctx)
	}
}

// IsPruningEnabled -
func (as *AccountsStub) IsPruningEnabled() bool {
	if as.IsPruningEnabledCalled != nil {
		return as.IsPruningEnabledCalled()
	}

	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (as *AccountsStub) IsInterfaceNil() bool {
	return as == nil
}

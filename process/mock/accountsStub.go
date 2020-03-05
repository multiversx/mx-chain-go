package mock

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go/data/state"
)

// AccountsStub -
type AccountsStub struct {
	CommitCalled             func() ([]byte, error)
	GetExistingAccountCalled func(addressContainer state.AddressContainer) (state.AccountHandler, error)
	HasAccountStateCalled    func(addressContainer state.AddressContainer) (bool, error)
	JournalLenCalled         func() int
	PutCodeCalled            func(accountHandler state.AccountHandler, code []byte) error
	RemoveAccountCalled      func(addressContainer state.AddressContainer) error
	RemoveCodeCalled         func(codeHash []byte) error
	RevertToSnapshotCalled   func(snapshot int) error
	SaveAccountStateCalled   func(acountWrapper state.AccountHandler) error
	SaveDataTrieCalled       func(acountWrapper state.AccountHandler) error
	RootHashCalled           func() ([]byte, error)
	RecreateTrieCalled       func(rootHash []byte) error
	PruneTrieCalled          func(rootHash []byte) error
	SnapshotStateCalled      func(rootHash []byte)
	SetStateCheckpointCalled func(rootHash []byte)
	CancelPruneCalled        func(rootHash []byte)
	IsPruningEnabledCalled   func() bool
	GetAllLeavesCalled       func(rootHash []byte) (map[string][]byte, error)
	LoadAccountCalled        func(container state.AddressContainer) (state.AccountHandler, error)
	SaveAccountCalled        func(account state.AccountHandler) error
}

func (as *AccountsStub) LoadAccount(address state.AddressContainer) (state.AccountHandler, error) {
	if as.LoadAccountCalled != nil {
		return as.LoadAccountCalled(address)
	}
	return nil, nil
}

func (as *AccountsStub) SaveAccount(account state.AccountHandler) error {
	if as.SaveAccountCalled != nil {
		return as.SaveAccountCalled(account)
	}
	return nil
}

// GetAllLeaves -
func (as *AccountsStub) GetAllLeaves(rootHash []byte) (map[string][]byte, error) {
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
func (as *AccountsStub) GetExistingAccount(addressContainer state.AddressContainer) (state.AccountHandler, error) {
	if as.GetExistingAccountCalled != nil {
		return as.GetExistingAccountCalled(addressContainer)
	}

	return nil, errNotImplemented
}

// HasAccount -
func (as *AccountsStub) HasAccount(addressContainer state.AddressContainer) (bool, error) {
	if as.HasAccountStateCalled != nil {
		return as.HasAccountStateCalled(addressContainer)
	}

	return false, errNotImplemented
}

// JournalLen -
func (as *AccountsStub) JournalLen() int {
	if as.JournalLenCalled != nil {
		return as.JournalLenCalled()
	}

	return 0
}

// PutCode -
func (as *AccountsStub) PutCode(accountHandler state.AccountHandler, code []byte) error {
	if as.PutCodeCalled != nil {
		return as.PutCodeCalled(accountHandler, code)
	}

	return errNotImplemented
}

// RemoveAccount -
func (as *AccountsStub) RemoveAccount(addressContainer state.AddressContainer) error {
	if as.RemoveAccountCalled != nil {
		return as.RemoveAccountCalled(addressContainer)
	}

	return errNotImplemented
}

// RemoveCode -
func (as *AccountsStub) RemoveCode(codeHash []byte) error {
	if as.RemoveCodeCalled != nil {
		return as.RemoveCodeCalled(codeHash)
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

// SaveJournalizedAccount -
func (as *AccountsStub) SaveJournalizedAccount(journalizedAccountHandler state.AccountHandler) error {
	if as.SaveAccountStateCalled != nil {
		return as.SaveAccountStateCalled(journalizedAccountHandler)
	}

	return errNotImplemented
}

// SaveDataTrie -
func (as *AccountsStub) SaveDataTrie(journalizedAccountHandler state.AccountHandler) error {
	if as.SaveDataTrieCalled != nil {
		return as.SaveDataTrieCalled(journalizedAccountHandler)
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
func (as *AccountsStub) PruneTrie(rootHash []byte) error {
	if as.PruneTrieCalled != nil {
		return as.PruneTrieCalled(rootHash)
	}

	return errNotImplemented
}

// CancelPrune -
func (as *AccountsStub) CancelPrune(rootHash []byte) {
	if as.CancelPruneCalled != nil {
		as.CancelPruneCalled(rootHash)
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

// IsInterfaceNil returns true if there is no value under the interface
func (as *AccountsStub) IsInterfaceNil() bool {
	return as == nil
}

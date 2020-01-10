package mock

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go/data/state"
)

type AccountsStub struct {
	AddJournalEntryCalled       func(je state.JournalEntry)
	CommitCalled                func() ([]byte, error)
	GetAccountWithJournalCalled func(addressContainer state.AddressContainer) (state.AccountHandler, error)
	GetExistingAccountCalled    func(addressContainer state.AddressContainer) (state.AccountHandler, error)
	HasAccountStateCalled       func(addressContainer state.AddressContainer) (bool, error)
	JournalLenCalled            func() int
	PutCodeCalled               func(accountHandler state.AccountHandler, code []byte) error
	RemoveAccountCalled         func(addressContainer state.AddressContainer) error
	RemoveCodeCalled            func(codeHash []byte) error
	RevertToSnapshotCalled      func(snapshot int) error
	SaveAccountStateCalled      func(acountWrapper state.AccountHandler) error
	SaveDataTrieCalled          func(acountWrapper state.AccountHandler) error
	RootHashCalled              func() ([]byte, error)
	RecreateTrieCalled          func(rootHash []byte) error
}

var errNotImplemented = errors.New("not implemented")

func (as *AccountsStub) AddJournalEntry(je state.JournalEntry) {
	if as.AddJournalEntryCalled != nil {
		as.AddJournalEntryCalled(je)
	}
}

func (as *AccountsStub) Commit() ([]byte, error) {
	if as.CommitCalled != nil {
		return as.CommitCalled()
	}

	return nil, errNotImplemented
}

func (as *AccountsStub) ClosePersister() error {
	return nil
}

func (as *AccountsStub) GetAccountWithJournal(addressContainer state.AddressContainer) (state.AccountHandler, error) {
	if as.GetAccountWithJournalCalled != nil {
		return as.GetAccountWithJournalCalled(addressContainer)
	}

	return nil, errNotImplemented
}

func (as *AccountsStub) GetExistingAccount(addressContainer state.AddressContainer) (state.AccountHandler, error) {
	if as.GetExistingAccountCalled != nil {
		return as.GetExistingAccountCalled(addressContainer)
	}

	return nil, errNotImplemented
}

func (as *AccountsStub) HasAccount(addressContainer state.AddressContainer) (bool, error) {
	if as.HasAccountStateCalled != nil {
		return as.HasAccountStateCalled(addressContainer)
	}

	return false, errNotImplemented
}

func (as *AccountsStub) JournalLen() int {
	if as.JournalLenCalled != nil {
		return as.JournalLenCalled()
	}

	return 0
}

func (as *AccountsStub) PutCode(accountHandler state.AccountHandler, code []byte) error {
	if as.PutCodeCalled != nil {
		return as.PutCodeCalled(accountHandler, code)
	}

	return errNotImplemented
}

func (as *AccountsStub) RemoveAccount(addressContainer state.AddressContainer) error {
	if as.RemoveAccountCalled != nil {
		return as.RemoveAccountCalled(addressContainer)
	}

	return errNotImplemented
}

func (as *AccountsStub) RemoveCode(codeHash []byte) error {
	if as.RemoveCodeCalled != nil {
		return as.RemoveCodeCalled(codeHash)
	}

	return errNotImplemented
}

func (as *AccountsStub) RevertToSnapshot(snapshot int) error {
	if as.RevertToSnapshotCalled != nil {
		return as.RevertToSnapshotCalled(snapshot)
	}

	return errNotImplemented
}

func (as *AccountsStub) SaveJournalizedAccount(journalizedAccountHandler state.AccountHandler) error {
	if as.SaveAccountStateCalled != nil {
		return as.SaveAccountStateCalled(journalizedAccountHandler)
	}

	return errNotImplemented
}

func (as *AccountsStub) SaveDataTrie(journalizedAccountHandler state.AccountHandler) error {
	if as.SaveDataTrieCalled != nil {
		return as.SaveDataTrieCalled(journalizedAccountHandler)
	}

	return errNotImplemented
}

func (as *AccountsStub) RootHash() ([]byte, error) {
	if as.RootHashCalled != nil {
		return as.RootHashCalled()
	}

	return nil, errNotImplemented
}

func (as *AccountsStub) RecreateTrie(rootHash []byte) error {
	if as.RecreateTrieCalled != nil {
		return as.RecreateTrieCalled(rootHash)
	}

	return errNotImplemented
}

// IsInterfaceNil returns true if there is no value under the interface
func (as *AccountsStub) IsInterfaceNil() bool {
	if as == nil {
		return true
	}
	return false
}

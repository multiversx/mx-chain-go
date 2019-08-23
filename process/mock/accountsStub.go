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

func NewAccountsStub() *AccountsStub {
	return &AccountsStub{}
}

func (aam *AccountsStub) AddJournalEntry(je state.JournalEntry) {
	if aam.AddJournalEntryCalled != nil {
		aam.AddJournalEntryCalled(je)
	}
}

func (aam *AccountsStub) Commit() ([]byte, error) {
	if aam.CommitCalled != nil {
		return aam.CommitCalled()
	}

	return nil, errNotImplemented
}

func (aam *AccountsStub) GetAccountWithJournal(addressContainer state.AddressContainer) (state.AccountHandler, error) {
	if aam.GetAccountWithJournalCalled != nil {
		return aam.GetAccountWithJournalCalled(addressContainer)
	}

	return nil, errNotImplemented
}

func (aam *AccountsStub) GetExistingAccount(addressContainer state.AddressContainer) (state.AccountHandler, error) {
	if aam.GetExistingAccountCalled != nil {
		return aam.GetExistingAccountCalled(addressContainer)
	}

	return nil, errNotImplemented
}

func (aam *AccountsStub) HasAccount(addressContainer state.AddressContainer) (bool, error) {
	if aam.HasAccountStateCalled != nil {
		return aam.HasAccountStateCalled(addressContainer)
	}

	return false, errNotImplemented
}

func (aam *AccountsStub) JournalLen() int {
	if aam.JournalLenCalled != nil {
		return aam.JournalLenCalled()
	}

	return 0
}

func (aam *AccountsStub) PutCode(accountHandler state.AccountHandler, code []byte) error {
	if aam.PutCodeCalled != nil {
		return aam.PutCodeCalled(accountHandler, code)
	}

	return errNotImplemented
}

func (aam *AccountsStub) RemoveAccount(addressContainer state.AddressContainer) error {
	if aam.RemoveAccountCalled != nil {
		return aam.RemoveAccountCalled(addressContainer)
	}

	return errNotImplemented
}

func (aam *AccountsStub) RemoveCode(codeHash []byte) error {
	if aam.RemoveCodeCalled != nil {
		return aam.RemoveCodeCalled(codeHash)
	}

	return errNotImplemented
}

func (aam *AccountsStub) RevertToSnapshot(snapshot int) error {
	if aam.RevertToSnapshotCalled != nil {
		return aam.RevertToSnapshotCalled(snapshot)
	}

	return errNotImplemented
}

func (aam *AccountsStub) SaveJournalizedAccount(journalizedAccountHandler state.AccountHandler) error {
	if aam.SaveAccountStateCalled != nil {
		return aam.SaveAccountStateCalled(journalizedAccountHandler)
	}

	return errNotImplemented
}

func (aam *AccountsStub) SaveDataTrie(journalizedAccountHandler state.AccountHandler) error {
	if aam.SaveDataTrieCalled != nil {
		return aam.SaveDataTrieCalled(journalizedAccountHandler)
	}

	return errNotImplemented
}

func (aam *AccountsStub) RootHash() ([]byte, error) {
	if aam.RootHashCalled != nil {
		return aam.RootHashCalled()
	}

	return nil, errNotImplemented
}

func (aam *AccountsStub) RecreateTrie(rootHash []byte) error {
	if aam.RecreateTrieCalled != nil {
		return aam.RecreateTrieCalled(rootHash)
	}

	return errNotImplemented
}

// IsInterfaceNil returns true if there is no value under the interface
func (aam *AccountsStub) IsInterfaceNil() bool {
	if aam == nil {
		return true
	}
	return false
}

package mock

import (
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

func (aam *AccountsStub) AddJournalEntry(je state.JournalEntry) {
	aam.AddJournalEntryCalled(je)
}

func (aam *AccountsStub) Commit() ([]byte, error) {
	return aam.CommitCalled()
}

func (aam *AccountsStub) GetAccountWithJournal(addressContainer state.AddressContainer) (state.AccountHandler, error) {
	return aam.GetAccountWithJournalCalled(addressContainer)
}

func (aam *AccountsStub) GetExistingAccount(addressContainer state.AddressContainer) (state.AccountHandler, error) {
	return aam.GetExistingAccountCalled(addressContainer)
}

func (aam *AccountsStub) HasAccount(addressContainer state.AddressContainer) (bool, error) {
	return aam.HasAccountStateCalled(addressContainer)
}

func (aam *AccountsStub) JournalLen() int {
	return aam.JournalLenCalled()
}

func (aam *AccountsStub) PutCode(accountHandler state.AccountHandler, code []byte) error {
	return aam.PutCodeCalled(accountHandler, code)
}

func (aam *AccountsStub) RemoveAccount(addressContainer state.AddressContainer) error {
	return aam.RemoveAccountCalled(addressContainer)
}

func (aam *AccountsStub) RemoveCode(codeHash []byte) error {
	return aam.RemoveCodeCalled(codeHash)
}

func (aam *AccountsStub) RevertToSnapshot(snapshot int) error {
	return aam.RevertToSnapshotCalled(snapshot)
}

func (aam *AccountsStub) SaveJournalizedAccount(journalizedAccountHandler state.AccountHandler) error {
	return aam.SaveAccountStateCalled(journalizedAccountHandler)
}

func (aam *AccountsStub) SaveDataTrie(journalizedAccountHandler state.AccountHandler) error {
	return aam.SaveDataTrieCalled(journalizedAccountHandler)
}

func (aam *AccountsStub) RootHash() ([]byte, error) {
	return aam.RootHashCalled()
}

func (aam *AccountsStub) RecreateTrie(rootHash []byte) error {
	return aam.RecreateTrieCalled(rootHash)
}

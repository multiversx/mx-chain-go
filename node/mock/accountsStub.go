package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type AccountsStub struct {
	AddJournalEntryCalled       func(je state.JournalEntry)
	CommitCalled                func() ([]byte, error)
	GetAccountWithJournalCalled func(addressContainer state.AddressContainer) (state.AccountWrapper, error)
	GetExistingAccountCalled    func(addressContainer state.AddressContainer) (state.AccountWrapper, error)
	HasAccountStateCalled       func(addressContainer state.AddressContainer) (bool, error)
	JournalLenCalled            func() int
	PutCodeCalled               func(accountWrapper state.AccountWrapper, code []byte) error
	RemoveAccountCalled         func(addressContainer state.AddressContainer) error
	RemoveCodeCalled            func(codeHash []byte) error
	RetrieveDataTrieCalled      func(acountWrapper state.AccountWrapper) error
	RevertToSnapshotCalled      func(snapshot int) error
	SaveAccountStateCalled      func(acountWrapper state.AccountWrapper) error
	SaveDataTrieCalled          func(acountWrapper state.AccountWrapper) error
	RootHashCalled              func() []byte
	RecreateTrieCalled          func(rootHash []byte) error
}

func NewAccountsStub() *AccountsStub {
	return &AccountsStub{}
}

func (aam *AccountsStub) AddJournalEntry(je state.JournalEntry) {
	aam.AddJournalEntryCalled(je)
}

func (aam *AccountsStub) Commit() ([]byte, error) {
	return aam.CommitCalled()
}

func (aam *AccountsStub) GetAccountWithJournal(addressContainer state.AddressContainer) (state.AccountWrapper, error) {
	return aam.GetAccountWithJournalCalled(addressContainer)
}

func (aam *AccountsStub) GetExistingAccount(addressContainer state.AddressContainer) (state.AccountWrapper, error) {
	return aam.GetExistingAccountCalled(addressContainer)
}

func (aam *AccountsStub) HasAccount(addressContainer state.AddressContainer) (bool, error) {
	return aam.HasAccountStateCalled(addressContainer)
}

func (aam *AccountsStub) JournalLen() int {
	return aam.JournalLenCalled()
}

func (aam *AccountsStub) PutCode(accountWrapper state.AccountWrapper, code []byte) error {
	return aam.PutCodeCalled(accountWrapper, code)
}

func (aam *AccountsStub) RemoveAccount(addressContainer state.AddressContainer) error {
	return aam.RemoveAccountCalled(addressContainer)
}

func (aam *AccountsStub) RemoveCode(codeHash []byte) error {
	return aam.RemoveCodeCalled(codeHash)
}

func (aam *AccountsStub) LoadDataTrie(accountWrapper state.AccountWrapper) error {
	return aam.RetrieveDataTrieCalled(accountWrapper)
}

func (aam *AccountsStub) RevertToSnapshot(snapshot int) error {
	return aam.RevertToSnapshotCalled(snapshot)
}

func (aam *AccountsStub) SaveJournalizedAccount(journalizedAccountWrapper state.AccountWrapper) error {
	return aam.SaveAccountStateCalled(journalizedAccountWrapper)
}

func (aam *AccountsStub) SaveDataTrie(journalizedAccountWrapper state.AccountWrapper) error {
	return aam.SaveDataTrieCalled(journalizedAccountWrapper)
}

func (aam *AccountsStub) RootHash() []byte {
	return aam.RootHashCalled()
}

func (aam *AccountsStub) RecreateTrie(rootHash []byte) error {
	return aam.RecreateTrieCalled(rootHash)
}

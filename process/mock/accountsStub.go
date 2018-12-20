package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type AccountsStub struct {
	AddJournalEntryCalled        func(je state.JournalEntry)
	CommitCalled                 func() ([]byte, error)
	GetJournalizedAccountCalled  func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error)
	HasAccountCalled             func(addressContainer state.AddressContainer) (bool, error)
	JournalLenCalled             func() int
	PutCodeCalled                func(journalizedAccountWrapper state.JournalizedAccountWrapper, code []byte) error
	RemoveAccountCalled          func(addressContainer state.AddressContainer) error
	RemoveCodeCalled             func(codeHash []byte) error
	LoadDataTrieCalled           func(journalizedAccountWrapper state.JournalizedAccountWrapper) error
	RevertToSnapshotCalled       func(snapshot int) error
	SaveJournalizedAccountCalled func(journalizedAccountWrapper state.JournalizedAccountWrapper) error
	SaveDataCalled               func(journalizedAccountWrapper state.JournalizedAccountWrapper) error
	RootHashCalled               func() []byte
	RecreateTrieCalled           func(rootHash []byte) error
}

func (as *AccountsStub) AddJournalEntry(je state.JournalEntry) {
	as.AddJournalEntryCalled(je)
}

func (as *AccountsStub) Commit() ([]byte, error) {
	return as.CommitCalled()
}

func (as *AccountsStub) GetJournalizedAccount(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
	return as.GetJournalizedAccountCalled(addressContainer)
}

func (as *AccountsStub) HasAccount(addressContainer state.AddressContainer) (bool, error) {
	return as.HasAccountCalled(addressContainer)
}

func (as *AccountsStub) JournalLen() int {
	return as.JournalLenCalled()
}

func (as *AccountsStub) PutCode(journalizedAccountWrapper state.JournalizedAccountWrapper, code []byte) error {
	return as.PutCodeCalled(journalizedAccountWrapper, code)
}

func (as *AccountsStub) RemoveAccount(addressContainer state.AddressContainer) error {
	return as.RemoveAccountCalled(addressContainer)
}

func (as *AccountsStub) RemoveCode(codeHash []byte) error {
	return as.RemoveCodeCalled(codeHash)
}

func (as *AccountsStub) LoadDataTrie(journalizedAccountWrapper state.JournalizedAccountWrapper) error {
	return as.LoadDataTrieCalled(journalizedAccountWrapper)
}

func (as *AccountsStub) RevertToSnapshot(snapshot int) error {
	return as.RevertToSnapshotCalled(snapshot)
}

func (as *AccountsStub) SaveJournalizedAccount(journalizedAccountWrapper state.JournalizedAccountWrapper) error {
	return as.SaveJournalizedAccountCalled(journalizedAccountWrapper)
}

func (as *AccountsStub) SaveData(journalizedAccountWrapper state.JournalizedAccountWrapper) error {
	return as.SaveDataCalled(journalizedAccountWrapper)
}

func (as *AccountsStub) RootHash() []byte {
	return as.RootHashCalled()
}

func (as *AccountsStub) RecreateTrie(rootHash []byte) error {
	return as.RecreateTrieCalled(rootHash)
}

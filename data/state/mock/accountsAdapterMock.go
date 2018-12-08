package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type AccountsAdapterMock struct {
	AddJournalEntryCalled       func(je state.JournalEntry)
	CommitCalled                func() ([]byte, error)
	GetJournalizedAccountCalled func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error)
	HasAccountStateCalled       func(addressContainer state.AddressContainer) (bool, error)
	JournalLenCalled            func() int
	PutCodeCalled               func(accountWrapper state.JournalizedAccountWrapper, code []byte) error
	RemoveAccountCalled         func(addressContainer state.AddressContainer) error
	RemoveCodeCalled            func(codeHash []byte) error
	RetrieveDataTrieCalled      func(acountWrapper state.JournalizedAccountWrapper) error
	RevertToSnapshotCalled      func(snapshot int) error
	SaveAccountStateCalled      func(acountWrapper state.JournalizedAccountWrapper) error
	SaveDataCalled              func(acountWrapper state.JournalizedAccountWrapper) error
	RootHashCalled              func() []byte
	RecreateTrieCalled          func(rootHash []byte) error
}

func NewAccountsAdapterMock() *AccountsAdapterMock {
	return &AccountsAdapterMock{}
}

func (aam *AccountsAdapterMock) AddJournalEntry(je state.JournalEntry) {
	aam.AddJournalEntryCalled(je)
}

func (aam *AccountsAdapterMock) Commit() ([]byte, error) {
	return aam.CommitCalled()
}

func (aam *AccountsAdapterMock) GetJournalizedAccount(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
	return aam.GetJournalizedAccountCalled(addressContainer)
}

func (aam *AccountsAdapterMock) HasAccount(addressContainer state.AddressContainer) (bool, error) {
	return aam.HasAccountStateCalled(addressContainer)
}

func (aam *AccountsAdapterMock) JournalLen() int {
	return aam.JournalLenCalled()
}

func (aam *AccountsAdapterMock) PutCode(accountWrapper state.JournalizedAccountWrapper, code []byte) error {
	return aam.PutCodeCalled(accountWrapper, code)
}

func (aam *AccountsAdapterMock) RemoveAccount(addressContainer state.AddressContainer) error {
	return aam.RemoveAccountCalled(addressContainer)
}

func (aam *AccountsAdapterMock) RemoveCode(codeHash []byte) error {
	return aam.RemoveCodeCalled(codeHash)
}

func (aam *AccountsAdapterMock) LoadDataTrie(accountWrapper state.JournalizedAccountWrapper) error {
	return aam.RetrieveDataTrieCalled(accountWrapper)
}

func (aam *AccountsAdapterMock) RevertToSnapshot(snapshot int) error {
	return aam.RevertToSnapshotCalled(snapshot)
}

func (aam *AccountsAdapterMock) SaveJournalizedAccount(journalizedAccountWrapper state.JournalizedAccountWrapper) error {
	return aam.SaveAccountStateCalled(journalizedAccountWrapper)
}

func (aam *AccountsAdapterMock) SaveData(journalizedAccountWrapper state.JournalizedAccountWrapper) error {
	return aam.SaveDataCalled(journalizedAccountWrapper)
}

func (aam *AccountsAdapterMock) RootHash() []byte {
	return aam.RootHashCalled()
}

func (aam *AccountsAdapterMock) RecreateTrie(rootHash []byte) error {
	return aam.RecreateTrieCalled(rootHash)
}

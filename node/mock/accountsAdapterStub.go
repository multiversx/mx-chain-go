package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type AccountsAdapterStub struct {
	AddJournalEntryHandler        func(je state.JournalEntry)
	CommitHandler                 func() ([]byte, error)
	GetJournalizedAccountHandler  func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error)
	GetExistingAccountHandler     func(addressContainer state.AddressContainer) (state.AccountWrapper, error)
	HasAccountHandler             func(addressContainer state.AddressContainer) (bool, error)
	JournalLenHandler             func() int
	PutCodeHandler                func(journalizedAccountWrapper state.JournalizedAccountWrapper, code []byte) error
	RemoveAccountHandler          func(addressContainer state.AddressContainer) error
	RemoveCodeHandler             func(codeHash []byte) error
	LoadDataTrieHandler           func(accountWrapper state.AccountWrapper) error
	RevertToSnapshotHandler       func(snapshot int) error
	SaveJournalizedAccountHandler func(journalizedAccountWrapper state.JournalizedAccountWrapper) error
	SaveDataHandler               func(journalizedAccountWrapper state.JournalizedAccountWrapper) error
	RootHashHandler               func() []byte
	RecreateTrieHandler           func(rootHash []byte) error
}

func (ad AccountsAdapterStub) AddJournalEntry(je state.JournalEntry) {
	ad.AddJournalEntryHandler(je)
}

func (ad AccountsAdapterStub) Commit() ([]byte, error) {
	return ad.CommitHandler()
}

func (ad AccountsAdapterStub) GetJournalizedAccount(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
	return ad.GetJournalizedAccountHandler(addressContainer)
}

func (ad AccountsAdapterStub) GetExistingAccount(addressContainer state.AddressContainer) (state.AccountWrapper, error) {
	return ad.GetExistingAccountHandler(addressContainer)
}

func (ad AccountsAdapterStub) HasAccount(addressContainer state.AddressContainer) (bool, error) {
	return ad.HasAccountHandler(addressContainer)
}

func (ad AccountsAdapterStub) JournalLen() int {
	return ad.JournalLenHandler()
}
func (ad AccountsAdapterStub) PutCode(journalizedAccountWrapper state.JournalizedAccountWrapper, code []byte) error {
	return ad.PutCodeHandler(journalizedAccountWrapper, code)
}
func (ad AccountsAdapterStub) RemoveAccount(addressContainer state.AddressContainer) error {
	return ad.RemoveAccountHandler(addressContainer)
}
func (ad AccountsAdapterStub) RemoveCode(codeHash []byte) error {
	return ad.RemoveCodeHandler(codeHash)
}
func (ad AccountsAdapterStub) LoadDataTrie(accountWrapper state.AccountWrapper) error {
	return ad.LoadDataTrieHandler(accountWrapper)
}
func (ad AccountsAdapterStub) RevertToSnapshot(snapshot int) error {
	return ad.RevertToSnapshotHandler(snapshot)
}
func (ad AccountsAdapterStub) SaveJournalizedAccount(journalizedAccountWrapper state.JournalizedAccountWrapper) error {
	return ad.SaveJournalizedAccountHandler(journalizedAccountWrapper)
}
func (ad AccountsAdapterStub) SaveData(journalizedAccountWrapper state.JournalizedAccountWrapper) error {
	return ad.SaveDataHandler(journalizedAccountWrapper)
}
func (ad AccountsAdapterStub) RootHash() []byte {
	return ad.RootHashHandler()
}
func (ad AccountsAdapterStub) RecreateTrie(rootHash []byte) error {
	return ad.RecreateTrie(rootHash)
}

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
	PruneTrieCalled             func(rootHash []byte) error
	SnapshotStateCalled         func(rootHash []byte)
	SetStateCheckpointCalled    func(rootHash []byte)
	CancelPruneCalled           func(rootHash []byte)
	IsPruningEnabledCalled      func() bool
}

func (as *AccountsStub) ClosePersister() error {
	return nil
}

func (as *AccountsStub) AddJournalEntry(je state.JournalEntry) {
	as.AddJournalEntryCalled(je)
}

func (as *AccountsStub) Commit() ([]byte, error) {
	return as.CommitCalled()
}

func (as *AccountsStub) GetAccountWithJournal(addressContainer state.AddressContainer) (state.AccountHandler, error) {
	return as.GetAccountWithJournalCalled(addressContainer)
}

func (as *AccountsStub) GetExistingAccount(addressContainer state.AddressContainer) (state.AccountHandler, error) {
	return as.GetExistingAccountCalled(addressContainer)
}

func (as *AccountsStub) HasAccount(addressContainer state.AddressContainer) (bool, error) {
	return as.HasAccountStateCalled(addressContainer)
}

func (as *AccountsStub) JournalLen() int {
	return as.JournalLenCalled()
}

func (as *AccountsStub) PutCode(accountHandler state.AccountHandler, code []byte) error {
	return as.PutCodeCalled(accountHandler, code)
}

func (as *AccountsStub) RemoveAccount(addressContainer state.AddressContainer) error {
	return as.RemoveAccountCalled(addressContainer)
}

func (as *AccountsStub) RemoveCode(codeHash []byte) error {
	return as.RemoveCodeCalled(codeHash)
}

func (as *AccountsStub) RevertToSnapshot(snapshot int) error {
	return as.RevertToSnapshotCalled(snapshot)
}

func (as *AccountsStub) SaveJournalizedAccount(journalizedAccountHandler state.AccountHandler) error {
	return as.SaveAccountStateCalled(journalizedAccountHandler)
}

func (as *AccountsStub) SaveDataTrie(journalizedAccountHandler state.AccountHandler) error {
	return as.SaveDataTrieCalled(journalizedAccountHandler)
}

func (as *AccountsStub) RootHash() ([]byte, error) {
	return as.RootHashCalled()
}

func (as *AccountsStub) RecreateTrie(rootHash []byte) error {
	return as.RecreateTrieCalled(rootHash)
}

func (as *AccountsStub) PruneTrie(rootHash []byte) error {
	return as.PruneTrieCalled(rootHash)
}

func (as *AccountsStub) CancelPrune(rootHash []byte) {
	as.CancelPruneCalled(rootHash)
}

func (as *AccountsStub) SnapshotState(rootHash []byte) {
	as.SnapshotStateCalled(rootHash)
}

func (as *AccountsStub) SetStateCheckpoint(rootHash []byte) {
	as.SetStateCheckpointCalled(rootHash)
}

func (as *AccountsStub) IsPruningEnabled() bool {
	return as.IsPruningEnabledCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (as *AccountsStub) IsInterfaceNil() bool {
	return as == nil
}

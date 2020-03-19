package disabled

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

type accountsAdapter struct {
}

// NewAccountsAdapter returns a new instance of accountsAdapter
func NewAccountsAdapter() *accountsAdapter {
	return &accountsAdapter{}
}

func (a *accountsAdapter) GetAccountWithJournal(addressContainer state.AddressContainer) (state.AccountHandler, error) {
	return nil, nil
}

func (a *accountsAdapter) GetExistingAccount(addressContainer state.AddressContainer) (state.AccountHandler, error) {
	return nil, nil
}

func (a *accountsAdapter) HasAccount(addressContainer state.AddressContainer) (bool, error) {
	return false, nil
}

func (a *accountsAdapter) RemoveAccount(addressContainer state.AddressContainer) error {
	return nil
}

func (a *accountsAdapter) Commit() ([]byte, error) {
	return nil, nil
}

func (a *accountsAdapter) JournalLen() int {
	return 0
}

func (a *accountsAdapter) RevertToSnapshot(snapshot int) error {
	return nil
}

func (a *accountsAdapter) RootHash() ([]byte, error) {
	return nil, nil
}

func (a *accountsAdapter) RecreateTrie(rootHash []byte) error {
	return nil
}

func (a *accountsAdapter) PutCode(accountHandler state.AccountHandler, code []byte) error {
	return nil
}

func (a *accountsAdapter) RemoveCode(codeHash []byte) error {
	return nil
}

func (a *accountsAdapter) SaveDataTrie(accountHandler state.AccountHandler) error {
	return nil
}

func (a *accountsAdapter) PruneTrie(rootHash []byte, identifier data.TriePruningIdentifier) error {
	return nil
}

func (a *accountsAdapter) CancelPrune(rootHash []byte, identifier data.TriePruningIdentifier) {
	return
}

func (a *accountsAdapter) SnapshotState(rootHash []byte) {
	return
}

func (a *accountsAdapter) SetStateCheckpoint(rootHash []byte) {
	return
}

func (a *accountsAdapter) IsPruningEnabled() bool {
	return false
}

func (a *accountsAdapter) ClosePersister() error {
	return nil
}

func (a *accountsAdapter) GetAllLeaves(rootHash []byte) (map[string][]byte, error) {
	return nil, nil
}

func (a *accountsAdapter) RecreateAllTries(rootHash []byte) (map[string]data.Trie, error) {
	return nil, nil
}

func (a *accountsAdapter) IsInterfaceNil() bool {
	return a == nil
}

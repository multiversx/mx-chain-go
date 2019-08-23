package state

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

//------- BaseJournalEntryCreation

// BaseJournalEntryCreation creates a new account entry in the state trie
// through updater it can revert the created changes.
type BaseJournalEntryCreation struct {
	key     []byte
	updater Updater
}

// NewBaseJournalEntryCreation outputs a new BaseJournalEntry implementation used to revert an account creation
func NewBaseJournalEntryCreation(key []byte, updater Updater) (*BaseJournalEntryCreation, error) {
	if updater == nil || updater.IsInterfaceNil() {
		return nil, ErrNilUpdater
	}
	if len(key) == 0 {
		return nil, ErrNilOrEmptyKey
	}

	return &BaseJournalEntryCreation{
		key:     key,
		updater: updater,
	}, nil
}

// Revert applies undo operation
func (bjec *BaseJournalEntryCreation) Revert() (AccountHandler, error) {
	return nil, bjec.updater.Update(bjec.key, nil)
}

// IsInterfaceNil returns true if there is no value under the interface
func (bjec *BaseJournalEntryCreation) IsInterfaceNil() bool {
	if bjec == nil {
		return true
	}
	return false
}

//------- BaseJournalEntryCodeHash

// BaseJournalEntryCodeHash creates a code hash change in account
type BaseJournalEntryCodeHash struct {
	account     AccountHandler
	oldCodeHash []byte
}

// NewBaseJournalEntryCodeHash outputs a new BaseJournalEntry implementation used to save and revert a code hash change
func NewBaseJournalEntryCodeHash(account AccountHandler, oldCodeHash []byte) (*BaseJournalEntryCodeHash, error) {
	if account == nil || account.IsInterfaceNil() {
		return nil, ErrNilAccountHandler
	}

	return &BaseJournalEntryCodeHash{
		account:     account,
		oldCodeHash: oldCodeHash,
	}, nil
}

// Revert applies undo operation
func (bjech *BaseJournalEntryCodeHash) Revert() (AccountHandler, error) {
	bjech.account.SetCodeHash(bjech.oldCodeHash)

	return bjech.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bjech *BaseJournalEntryCodeHash) IsInterfaceNil() bool {
	if bjech == nil {
		return true
	}
	return false
}

//------- BaseJournalEntryRoot

// BaseJournalEntryRootHash creates an account's root hash change
type BaseJournalEntryRootHash struct {
	account     AccountHandler
	oldRootHash []byte
	oldTrie     data.Trie
}

// NewBaseJournalEntryRootHash outputs a new BaseJournalEntry used to save and revert an account's root hash change
func NewBaseJournalEntryRootHash(account AccountHandler, oldRootHash []byte, oldTrie data.Trie) (*BaseJournalEntryRootHash, error) {
	if account == nil || account.IsInterfaceNil() {
		return nil, ErrNilAccountHandler
	}

	return &BaseJournalEntryRootHash{
		account:     account,
		oldRootHash: oldRootHash,
		oldTrie:     oldTrie,
	}, nil
}

// Revert applies undo operation
func (bjer *BaseJournalEntryRootHash) Revert() (AccountHandler, error) {
	bjer.account.SetRootHash(bjer.oldRootHash)
	bjer.account.SetDataTrie(bjer.oldTrie)

	return bjer.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bjer *BaseJournalEntryRootHash) IsInterfaceNil() bool {
	if bjer == nil {
		return true
	}
	return false
}

//------- BaseJournalEntryData

// BaseJournalEntryData is used to mark an account's data change
type BaseJournalEntryData struct {
	trie    data.Trie
	account AccountHandler
}

// NewBaseJournalEntryData outputs a new BaseJournalEntry implementation used to keep track of data change.
// The revert will practically empty the dirty data map
func NewBaseJournalEntryData(account AccountHandler, trie data.Trie) (*BaseJournalEntryData, error) {
	if account == nil || account.IsInterfaceNil() {
		return nil, ErrNilAccountHandler
	}

	return &BaseJournalEntryData{
		account: account,
		trie:    trie,
	}, nil
}

// Revert will empty the dirtyData map from AccountState
func (bjed *BaseJournalEntryData) Revert() (AccountHandler, error) {
	dataTrieTracker := bjed.account.DataTrieTracker()
	if dataTrieTracker != nil {
		bjed.account.DataTrieTracker().ClearDataCaches()
	}

	return nil, nil
}

// Trie returns the referenced PatriciaMerkelTree for committing the changes
func (bjed *BaseJournalEntryData) Trie() data.Trie {
	return bjed.trie
}

// IsInterfaceNil returns true if there is no value under the interface
func (bjed *BaseJournalEntryData) IsInterfaceNil() bool {
	if bjed == nil {
		return true
	}
	return false
}

//------- BaseJournalEntryNonce

// BaseJournalEntryNonce is used to revert a nonce change
type BaseJournalEntryNonce struct {
	account  AccountHandler
	oldNonce uint64
}

// NewBaseJournalEntryNonce outputs a new JournalEntry implementation used to revert a nonce change
func NewBaseJournalEntryNonce(account AccountHandler, oldNonce uint64) (*BaseJournalEntryNonce, error) {
	if account == nil || account.IsInterfaceNil() {
		return nil, ErrNilAccountHandler
	}

	return &BaseJournalEntryNonce{
		account:  account,
		oldNonce: oldNonce,
	}, nil
}

// Revert applies undo operation
func (bjen *BaseJournalEntryNonce) Revert() (AccountHandler, error) {
	bjen.account.SetNonce(bjen.oldNonce)

	return bjen.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bjen *BaseJournalEntryNonce) IsInterfaceNil() bool {
	if bjen == nil {
		return true
	}
	return false
}

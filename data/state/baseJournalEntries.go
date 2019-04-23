package state

import "github.com/ElrondNetwork/elrond-go-sandbox/data/trie"

//------- BaseJournalEntryCreation

// BaseJournalEntryCreation is used to revert an account creation
type BaseJournalEntryCreation struct {
	key     []byte
	updater Updater
}

// NewBaseJournalEntryCreation outputs a new BaseJournalEntry implementation used to revert an account creation
func NewBaseJournalEntryCreation(key []byte, updater Updater) (*BaseJournalEntryCreation, error) {
	if updater == nil {
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
func (bjec *BaseJournalEntryCreation) Revert() (AccountWrapper, error) {
	return nil, bjec.updater.Update(bjec.key, nil)
}

//------- BaseJournalEntryCodeHash

// BaseJournalEntryCodeHash is used to revert a code hash change
type BaseJournalEntryCodeHash struct {
	account     AccountWrapper
	oldCodeHash []byte
}

// NewBaseJournalEntryCodeHash outputs a new BaseJournalEntry implementation used to revert a code hash change
func NewBaseJournalEntryCodeHash(account AccountWrapper, oldCodeHash []byte) (*BaseJournalEntryCodeHash, error) {
	if account == nil {
		return nil, ErrNilAccountWrapper
	}

	return &BaseJournalEntryCodeHash{
		account:     account,
		oldCodeHash: oldCodeHash,
	}, nil
}

// Revert applies undo operation
func (bjech *BaseJournalEntryCodeHash) Revert() (AccountWrapper, error) {
	bjech.account.SetCodeHash(bjech.oldCodeHash)

	return bjech.account, nil
}

//------- BaseJournalEntryRoot

// BaseJournalEntryRootHash is used to revert an account's root hash change
type BaseJournalEntryRootHash struct {
	account     AccountWrapper
	oldRootHash []byte
}

// NewBaseJournalEntryRootHash outputs a new BaseJournalEntry implementation used to revert an account's root hash change
func NewBaseJournalEntryRootHash(account AccountWrapper, oldRootHash []byte) (*BaseJournalEntryRootHash, error) {
	if account == nil {
		return nil, ErrNilAccountWrapper
	}

	return &BaseJournalEntryRootHash{
		account:     account,
		oldRootHash: oldRootHash,
	}, nil
}

// Revert applies undo operation
func (bjer *BaseJournalEntryRootHash) Revert() (AccountWrapper, error) {
	bjer.account.SetRootHash(bjer.oldRootHash)

	return bjer.account, nil
}

//------- BaseJournalEntryData

// BaseJournalEntryData is used to mark an account's data change
type BaseJournalEntryData struct {
	trie    trie.PatriciaMerkelTree
	account AccountWrapper
}

// NewBaseJournalEntryData outputs a new BaseJournalEntry implementation used to keep track of data change.
// The revert will practically empty the dirty data map
func NewBaseJournalEntryData(account AccountWrapper, trie trie.PatriciaMerkelTree) (*BaseJournalEntryData, error) {
	if account == nil {
		return nil, ErrNilAccountWrapper
	}

	return &BaseJournalEntryData{
		account: account,
		trie:    trie,
	}, nil
}

// Revert will empty the dirtyData map from AccountState
func (bjed *BaseJournalEntryData) Revert() (AccountWrapper, error) {
	dataTrieTracker := bjed.account.DataTrieTracker()
	if dataTrieTracker != nil {
		bjed.account.DataTrieTracker().ClearDataCaches()
	}

	return nil, nil
}

// Trie returns the referenced PatriciaMerkelTree for committing the changes
func (jed *BaseJournalEntryData) Trie() trie.PatriciaMerkelTree {
	return jed.trie
}

package state

import "github.com/ElrondNetwork/elrond-go-sandbox/data/trie"

//------- BaseJournalEntryCreation

// BaseJournalEntryCreation creates a new account entry in the state trie
// through updater it can revert the created changes.
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
func (bjec *BaseJournalEntryCreation) Revert() (AccountHandler, error) {
	return nil, bjec.updater.Update(bjec.key, nil)
}

//------- BaseJournalEntryCodeHash

// BaseJournalEntryCodeHash creates a code hash change in account
type BaseJournalEntryCodeHash struct {
	account     AccountHandler
	oldCodeHash []byte
}

// NewBaseJournalEntryCodeHash outputs a new BaseJournalEntry implementation used to save and revert a code hash change
func NewBaseJournalEntryCodeHash(account AccountHandler, oldCodeHash []byte) (*BaseJournalEntryCodeHash, error) {
	if account == nil {
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

//------- BaseJournalEntryCode

// BaseJournalEntryCode creates a code hash change in account
type BaseJournalEntryCode struct {
	account AccountHandler
	oldCode []byte
}

// NewBaseJournalEntryCode outputs a new BaseJournalEntry implementation used to save and revert a code change
func NewBaseJournalEntryCode(account AccountHandler, oldCode []byte) (*BaseJournalEntryCode, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &BaseJournalEntryCode{
		account: account,
		oldCode: oldCode,
	}, nil
}

// Revert applies undo operation
func (bjech *BaseJournalEntryCode) Revert() (AccountHandler, error) {
	bjech.account.SetCode(bjech.oldCode)

	return bjech.account, nil
}

//------- BaseJournalEntryRoot

// BaseJournalEntryRootHash creates an account's root hash change
type BaseJournalEntryRootHash struct {
	account     AccountHandler
	oldRootHash []byte
}

// NewBaseJournalEntryRootHash outputs a new BaseJournalEntry used to save and revert an account's root hash change
func NewBaseJournalEntryRootHash(account AccountHandler, oldRootHash []byte) (*BaseJournalEntryRootHash, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &BaseJournalEntryRootHash{
		account:     account,
		oldRootHash: oldRootHash,
	}, nil
}

// Revert applies undo operation
func (bjer *BaseJournalEntryRootHash) Revert() (AccountHandler, error) {
	bjer.account.SetRootHash(bjer.oldRootHash)

	return bjer.account, nil
}

//------- BaseJournalEntryData

// BaseJournalEntryData is used to mark an account's data change
type BaseJournalEntryData struct {
	trie    trie.PatriciaMerkelTree
	account AccountHandler
}

// NewBaseJournalEntryData outputs a new BaseJournalEntry implementation used to keep track of data change.
// The revert will practically empty the dirty data map
func NewBaseJournalEntryData(account AccountHandler, trie trie.PatriciaMerkelTree) (*BaseJournalEntryData, error) {
	if account == nil {
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
func (bjed *BaseJournalEntryData) Trie() trie.PatriciaMerkelTree {
	return bjed.trie
}

//------- JournalEntryNonce

// JournalEntryNonce is used to revert a nonce change
type JournalEntryNonce struct {
	account  AccountHandler
	oldNonce uint64
}

// NewJournalEntryNonce outputs a new JournalEntry implementation used to revert a nonce change
func NewJournalEntryNonce(account AccountHandler, oldNonce uint64) (*JournalEntryNonce, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &JournalEntryNonce{
		account:  account,
		oldNonce: oldNonce,
	}, nil
}

// Revert applies undo operation
func (jen *JournalEntryNonce) Revert() (AccountHandler, error) {
	jen.account.SetNonce(jen.oldNonce)

	return jen.account, nil
}

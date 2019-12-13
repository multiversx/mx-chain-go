package state

import (
	"math/big"
)

//------- JournalEntryBalance

// JournalEntryBalance is used to revert a balance change
type JournalEntryBalance struct {
	account    *Account
	oldBalance *big.Int
}

// NewJournalEntryBalance outputs a new JournalEntry implementation used to revert a balance change
func NewJournalEntryBalance(account *Account, oldBalance *big.Int) (*JournalEntryBalance, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &JournalEntryBalance{
		account:    account,
		oldBalance: oldBalance,
	}, nil
}

// Revert applies undo operation
func (jeb *JournalEntryBalance) Revert() (AccountHandler, error) {
	jeb.account.Balance = jeb.oldBalance

	return jeb.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (jeb *JournalEntryBalance) IsInterfaceNil() bool {
	if jeb == nil {
		return true
	}
	return false
}

//------- JournalEntryDataTrieUpdates

// JournalEntryDataTrieUpdates stores all the updates done to the account's data trie,
// so it can be reverted in case of rollback
type JournalEntryDataTrieUpdates struct {
	trieUpdates map[string][]byte
	account     AccountHandler
}

// NewJournalEntryDataTrieUpdates outputs a new JournalEntryDataTrieUpdates implementation used to revert an account's data trie
func NewJournalEntryDataTrieUpdates(trieUpdates map[string][]byte, account AccountHandler) (*JournalEntryDataTrieUpdates, error) {
	if account == nil || account.IsInterfaceNil() {
		return nil, ErrNilUpdater
	}
	if len(trieUpdates) == 0 {
		return nil, ErrNilOrEmptyDataTrieUpdates
	}

	return &JournalEntryDataTrieUpdates{
		trieUpdates: trieUpdates,
		account:     account,
	}, nil
}

// Revert applies undo operation
func (jedtu *JournalEntryDataTrieUpdates) Revert() (AccountHandler, error) {
	for key := range jedtu.trieUpdates {
		err := jedtu.account.DataTrie().Update([]byte(key), jedtu.trieUpdates[key])
		if err != nil {
			return nil, err
		}
	}

	rootHash, err := jedtu.account.DataTrie().Root()
	if err != nil {
		return nil, err
	}

	jedtu.account.SetRootHash(rootHash)

	return jedtu.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (jedtu *JournalEntryDataTrieUpdates) IsInterfaceNil() bool {
	return jedtu == nil
}

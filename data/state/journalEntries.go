package state

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
)

type JournalEntryCode struct {
	codeHash []byte
	updater  Updater
}

func NewJournalEntryCode(codeHash []byte, updater Updater) (*JournalEntryCode, error) {
	return &JournalEntryCode{
		codeHash: codeHash,
		updater:  updater,
	}, nil
}

// Revert applies undo operation
func (jea *JournalEntryCode) Revert() (AccountHandler, error) {
	return nil, jea.updater.Update(jea.codeHash, nil)
}

// IsInterfaceNil returns true if there is no value under the interface
func (jea *JournalEntryCode) IsInterfaceNil() bool {
	return jea == nil
}

//------- JournalEntryBalance

type JournalEntryAccount struct {
	account AccountHandler
}

func NewJournalEntryAccount(account AccountHandler) (*JournalEntryAccount, error) {
	if check.IfNil(account) {
		return nil, ErrNilAccountHandler
	}

	return &JournalEntryAccount{
		account: account,
	}, nil
}

// Revert applies undo operation
func (jea *JournalEntryAccount) Revert() (AccountHandler, error) {
	return jea.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (jea *JournalEntryAccount) IsInterfaceNil() bool {
	return jea == nil
}

type JournalEntryAccountCreation struct {
	address []byte
	updater Updater
}

func NewJournalEntryAccountCreation(address []byte, updater Updater) (*JournalEntryAccountCreation, error) {
	return &JournalEntryAccountCreation{
		address: address,
		updater: updater,
	}, nil
}

// Revert applies undo operation
func (jea *JournalEntryAccountCreation) Revert() (AccountHandler, error) {
	return nil, jea.updater.Update(jea.address, nil)
}

// IsInterfaceNil returns true if there is no value under the interface
func (jea *JournalEntryAccountCreation) IsInterfaceNil() bool {
	return jea == nil
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
	if check.IfNil(account) {
		return nil, ErrNilAccountHandler
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

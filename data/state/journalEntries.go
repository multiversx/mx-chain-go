package state

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/check"
)

type journalEntryCode struct {
	codeHash []byte
	updater  Updater
}

// NewJournalEntryCode creates a new instance of JournalEntryCode
func NewJournalEntryCode(codeHash []byte, updater Updater) (*journalEntryCode, error) {
	if check.IfNil(updater) {
		return nil, ErrNilUpdater
	}
	if len(codeHash) == 0 {
		return nil, ErrInvalidHash
	}

	return &journalEntryCode{
		codeHash: codeHash,
		updater:  updater,
	}, nil
}

// Revert applies undo operation
func (jea *journalEntryCode) Revert() (AccountHandler, error) {
	return nil, jea.updater.Update(jea.codeHash, nil)
}

// IsInterfaceNil returns true if there is no value under the interface
func (jea *journalEntryCode) IsInterfaceNil() bool {
	return jea == nil
}

// JournalEntryAccount represents a journal entry for account fields change
type journalEntryAccount struct {
	account AccountHandler
}

// NewJournalEntryAccount creates a new instance of JournalEntryAccount
func NewJournalEntryAccount(account AccountHandler) (*journalEntryAccount, error) {
	if check.IfNil(account) {
		return nil, fmt.Errorf("%w in NewJournalEntryAccount", ErrNilAccountHandler)
	}

	return &journalEntryAccount{
		account: account,
	}, nil
}

// Revert applies undo operation
func (jea *journalEntryAccount) Revert() (AccountHandler, error) {
	return jea.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (jea *journalEntryAccount) IsInterfaceNil() bool {
	return jea == nil
}

// JournalEntryAccountCreation represents a journal entry for account creation
type journalEntryAccountCreation struct {
	address []byte
	updater Updater
}

// NewJournalEntryAccountCreation creates a new instance of JournalEntryAccountCreation
func NewJournalEntryAccountCreation(address []byte, updater Updater) (*journalEntryAccountCreation, error) {
	if check.IfNil(updater) {
		return nil, ErrNilUpdater
	}
	if len(address) == 0 {
		return nil, ErrInvalidAddressLength
	}

	return &journalEntryAccountCreation{
		address: address,
		updater: updater,
	}, nil
}

// Revert applies undo operation
func (jea *journalEntryAccountCreation) Revert() (AccountHandler, error) {
	return nil, jea.updater.Update(jea.address, nil)
}

// IsInterfaceNil returns true if there is no value under the interface
func (jea *journalEntryAccountCreation) IsInterfaceNil() bool {
	return jea == nil
}

// JournalEntryDataTrieUpdates stores all the updates done to the account's data trie,
// so it can be reverted in case of rollback
type journalEntryDataTrieUpdates struct {
	trieUpdates map[string][]byte
	account     baseAccountHandler
}

// NewJournalEntryDataTrieUpdates outputs a new JournalEntryDataTrieUpdates implementation used to revert an account's data trie
func NewJournalEntryDataTrieUpdates(trieUpdates map[string][]byte, account baseAccountHandler) (*journalEntryDataTrieUpdates, error) {
	if check.IfNil(account) {
		return nil, fmt.Errorf("%w in NewJournalEntryDataTrieUpdates", ErrNilAccountHandler)
	}
	if len(trieUpdates) == 0 {
		return nil, ErrNilOrEmptyDataTrieUpdates
	}

	return &journalEntryDataTrieUpdates{
		trieUpdates: trieUpdates,
		account:     account,
	}, nil
}

// Revert applies undo operation
func (jedtu *journalEntryDataTrieUpdates) Revert() (AccountHandler, error) {
	for key := range jedtu.trieUpdates {
		err := jedtu.account.DataTrie().Update([]byte(key), jedtu.trieUpdates[key])
		if err != nil {
			return nil, err
		}

		log.Trace("revert data trie update", "key", []byte(key), "val", jedtu.trieUpdates[key])
	}

	rootHash, err := jedtu.account.DataTrie().Root()
	if err != nil {
		return nil, err
	}

	jedtu.account.SetRootHash(rootHash)

	return jedtu.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (jedtu *journalEntryDataTrieUpdates) IsInterfaceNil() bool {
	return jedtu == nil
}

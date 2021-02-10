package state

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

type journalEntryCode struct {
	oldCodeEntry *CodeEntry
	oldCodeHash  []byte
	newCodeHash  []byte
	trie         Updater
	marshalizer  marshal.Marshalizer
}

// NewJournalEntryCode creates a new instance of JournalEntryCode
func NewJournalEntryCode(
	oldCodeEntry *CodeEntry,
	oldCodeHash []byte,
	newCodeHash []byte,
	trie Updater,
	marshalizer marshal.Marshalizer,
) (*journalEntryCode, error) {
	if check.IfNil(trie) {
		return nil, ErrNilUpdater
	}
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}

	return &journalEntryCode{
		oldCodeEntry: oldCodeEntry,
		oldCodeHash:  oldCodeHash,
		newCodeHash:  newCodeHash,
		trie:         trie,
		marshalizer:  marshalizer,
	}, nil
}

// Revert applies undo operation
func (jea *journalEntryCode) Revert() (AccountHandler, error) {
	if bytes.Equal(jea.oldCodeHash, jea.newCodeHash) {
		return nil, nil
	}

	err := jea.revertOldCodeEntry()
	if err != nil {
		return nil, err
	}

	err = jea.revertNewCodeEntry()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (jea *journalEntryCode) revertOldCodeEntry() error {
	if len(jea.oldCodeHash) == 0 {
		return nil
	}

	err := saveCodeEntry(jea.oldCodeHash, jea.oldCodeEntry, jea.trie, jea.marshalizer)
	if err != nil {
		return err
	}

	return nil
}

func (jea *journalEntryCode) revertNewCodeEntry() error {
	newCodeEntry, err := getCodeEntry(jea.newCodeHash, jea.trie, jea.marshalizer)
	if err != nil {
		return err
	}

	if newCodeEntry == nil {
		return nil
	}

	if newCodeEntry.NumReferences <= 1 {
		err = jea.trie.Update(jea.newCodeHash, nil)
		if err != nil {
			return err
		}

		return nil
	}

	newCodeEntry.NumReferences--
	err = saveCodeEntry(jea.newCodeHash, newCodeEntry, jea.trie, jea.marshalizer)
	if err != nil {
		return err
	}

	return nil
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

	rootHash, err := jedtu.account.DataTrie().RootHash()
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

// journalEntryDataTrieRemove cancels the eviction of the hashes from the data trie with the given root hash
type journalEntryDataTrieRemove struct {
	rootHash               []byte
	obsoleteDataTrieHashes map[string][][]byte
}

// NewJournalEntryDataTrieRemove outputs a new journalEntryDataTrieRemove implementation used to cancel
// the eviction of the hashes from the data trie with the given root hash
func NewJournalEntryDataTrieRemove(rootHash []byte, obsoleteDataTrieHashes map[string][][]byte) (*journalEntryDataTrieRemove, error) {
	if obsoleteDataTrieHashes == nil {
		return nil, fmt.Errorf("%w in NewJournalEntryDataTrieRemove", ErrNilMapOfHashes)
	}
	if len(rootHash) == 0 {
		return nil, ErrInvalidRootHash
	}

	return &journalEntryDataTrieRemove{
		rootHash:               rootHash,
		obsoleteDataTrieHashes: obsoleteDataTrieHashes,
	}, nil
}

// Revert applies undo operation
func (jedtr *journalEntryDataTrieRemove) Revert() (AccountHandler, error) {
	delete(jedtr.obsoleteDataTrieHashes, string(jedtr.rootHash))

	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (jedtr *journalEntryDataTrieRemove) IsInterfaceNil() bool {
	return jedtr == nil
}

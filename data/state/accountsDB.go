package state

import (
	bytes "bytes"
	"encoding/hex"
	"fmt"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// AccountsDB is the struct used for accessing accounts. This struct is concurrent safe.
type AccountsDB struct {
	mainTrie       data.Trie
	hasher         hashing.Hasher
	marshalizer    marshal.Marshalizer
	accountFactory AccountFactory

	dataTries TriesHolder
	entries   []JournalEntry
	mutOp     sync.RWMutex
}

var log = logger.GetOrCreate("state")

// NewAccountsDB creates a new account manager
func NewAccountsDB(
	trie data.Trie,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	accountFactory AccountFactory,
) (*AccountsDB, error) {
	if check.IfNil(trie) {
		return nil, ErrNilTrie
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(accountFactory) {
		return nil, ErrNilAccountFactory
	}

	return &AccountsDB{
		mainTrie:       trie,
		hasher:         hasher,
		marshalizer:    marshalizer,
		accountFactory: accountFactory,
		entries:        make([]JournalEntry, 0),
		mutOp:          sync.RWMutex{},
		dataTries:      NewDataTriesHolder(),
	}, nil
}

// SaveAccount saves in the trie all changes made to the account.
func (adb *AccountsDB) SaveAccount(account AccountHandler) error {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	if check.IfNil(account) {
		return ErrNilAccountHandler
	}

	oldAccount, err := adb.getAccount(account.AddressContainer())
	if err != nil {
		return err
	}

	var entry JournalEntry
	if check.IfNil(oldAccount) {
		entry, err = NewJournalEntryAccountCreation(account.AddressContainer().Bytes(), adb.mainTrie)
		if err != nil {
			return err
		}
		adb.journalize(entry)
	} else {
		entry, err = NewJournalEntryAccount(oldAccount)
		if err != nil {
			return err
		}
		adb.journalize(entry)
	}

	baseAcc, ok := account.(baseAccountHandler)
	if ok {
		err = adb.saveCode(baseAcc)
		if err != nil {
			return err
		}

		err = adb.saveDataTrie(baseAcc)
		if err != nil {
			return err
		}
	}

	return adb.saveAccountToTrie(account)
}

func (adb *AccountsDB) saveCode(accountHandler baseAccountHandler) error {
	//TODO enable code pruning
	actualCode := accountHandler.GetCode()
	if len(actualCode) == 0 {
		return nil
	}

	fmt.Println("current code hash", accountHandler.GetCodeHash())

	actualCodeHash := adb.hasher.Compute(string(actualCode))
	//currentCodeHash := accountHandler.GetCodeHash()
	currentCode, err := adb.mainTrie.Get(actualCodeHash)
	if err != nil {
		return err
	}

	if len(currentCode) > 0 {
		if bytes.Equal(actualCode, currentCode) {
			fmt.Println("SAME CODE")
			// TODO: Explain why?
			accountHandler.SetCodeHash(actualCodeHash)
			return nil
		} else {
			fmt.Println("NOT SAME CODE")
			// // Update

			// // First delete current code
			// entry, err := NewJournalEntryCode(currentCodeHash, currentCode, adb.mainTrie)
			// if err != nil {
			// 	return err
			// }
			// adb.journalize(entry)

			// err = adb.mainTrie.Update(currentCodeHash, nil)
			// if err != nil {
			// 	return err
			// }

			// // Then set new code
			// entry, err = NewJournalEntryCode(actualCodeHash, nil, adb.mainTrie)
			// if err != nil {
			// 	return err
			// }
			// adb.journalize(entry)

			// err = adb.mainTrie.Update(actualCodeHash, actualCode)
			// if err != nil {
			// 	return err
			// }

			return nil
		}
	}

	//append a journal entry as the code needs to be inserted in the trie
	entry, err := NewJournalEntryCode(actualCodeHash, nil, adb.mainTrie)
	if err != nil {
		return err
	}
	adb.journalize(entry)

	//log.Trace("accountsDB.saveCode: mainTrie.Update()", "codeHash", actualCodeHash, "len(code)", len(actualCode))
	err = adb.mainTrie.Update(actualCodeHash, actualCode)
	if err != nil {
		return err
	}

	accountHandler.SetCodeHash(actualCodeHash)
	return nil
}

// LoadDataTrie retrieves and saves the SC data inside accountHandler object.
// Errors if something went wrong
func (adb *AccountsDB) loadDataTrie(accountHandler baseAccountHandler) error {
	if len(accountHandler.GetRootHash()) == 0 {
		return nil
	}

	dataTrie := adb.dataTries.Get(accountHandler.AddressContainer().Bytes())
	if dataTrie != nil {
		accountHandler.SetDataTrie(dataTrie)
		return nil
	}

	dataTrie, err := adb.mainTrie.Recreate(accountHandler.GetRootHash())
	if err != nil {
		return NewErrMissingTrie(accountHandler.GetRootHash())
	}

	accountHandler.SetDataTrie(dataTrie)
	adb.dataTries.Put(accountHandler.AddressContainer().Bytes(), dataTrie)
	return nil
}

// SaveDataTrie is used to save the data trie (not committing it) and to recompute the new Root value
// If data is not dirtied, method will not create its JournalEntries to keep track of data modification
func (adb *AccountsDB) saveDataTrie(accountHandler baseAccountHandler) error {
	if check.IfNil(accountHandler.DataTrieTracker()) {
		return ErrNilTrackableDataTrie
	}
	if len(accountHandler.DataTrieTracker().DirtyData()) == 0 {
		return nil
	}

	log.Trace("accountsDB.SaveDataTrie",
		"address", hex.EncodeToString(accountHandler.AddressContainer().Bytes()),
		"nonce", accountHandler.GetNonce(),
	)

	if check.IfNil(accountHandler.DataTrie()) {
		newDataTrie, err := adb.mainTrie.Recreate(make([]byte, 0))
		if err != nil {
			return err
		}

		accountHandler.SetDataTrie(newDataTrie)
		adb.dataTries.Put(accountHandler.AddressContainer().Bytes(), newDataTrie)
	}

	trackableDataTrie := accountHandler.DataTrieTracker()
	dataTrie := trackableDataTrie.DataTrie()
	oldValues := make(map[string][]byte)

	for k, v := range trackableDataTrie.DirtyData() {
		//TODO use trackableDataTrie.originalData() instead of getting from the trie
		val, err := dataTrie.Get([]byte(k))
		if err != nil {
			return err
		}

		oldValues[k] = val

		err = dataTrie.Update([]byte(k), v)
		if err != nil {
			return err
		}
	}

	entry, err := NewJournalEntryDataTrieUpdates(oldValues, accountHandler)
	if err != nil {
		return err
	}
	adb.journalize(entry)

	rootHash, err := trackableDataTrie.DataTrie().Root()
	if err != nil {
		return err
	}

	accountHandler.SetRootHash(rootHash)
	trackableDataTrie.ClearDataCaches()

	return nil
}

func (adb *AccountsDB) saveAccountToTrie(accountHandler AccountHandler) error {
	log.Trace("accountsDB.saveAccountToTrie",
		"address", hex.EncodeToString(accountHandler.AddressContainer().Bytes()),
		"nonce", accountHandler.GetNonce(),
	)

	//pass the reference to marshalizer, otherwise it will fail marshalizing balance
	buff, err := adb.marshalizer.Marshal(accountHandler)
	if err != nil {
		return err
	}

	return adb.mainTrie.Update(accountHandler.AddressContainer().Bytes(), buff)
}

// RemoveAccount removes the account data from underlying trie.
// It basically calls Update with empty slice
func (adb *AccountsDB) RemoveAccount(addressContainer AddressContainer) error {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	if check.IfNil(addressContainer) {
		return fmt.Errorf("%w in RemoveAccount", ErrNilAddressContainer)
	}

	acnt, err := adb.getAccount(addressContainer)
	if err != nil {
		return err
	}

	entry, err := NewJournalEntryAccount(acnt)
	if err != nil {
		return err
	}
	adb.journalize(entry)

	log.Trace("accountsDB.RemoveAccount",
		"address", hex.EncodeToString(addressContainer.Bytes()),
	)

	return adb.mainTrie.Update(addressContainer.Bytes(), make([]byte, 0))
}

// LoadAccount fetches the account based on the address. Creates an empty account if the account is missing.
func (adb *AccountsDB) LoadAccount(addressContainer AddressContainer) (AccountHandler, error) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	if check.IfNil(addressContainer) {
		return nil, fmt.Errorf("%w in LoadAccount", ErrNilAddressContainer)
	}

	log.Trace("accountsDB.LoadAccount",
		"address", hex.EncodeToString(addressContainer.Bytes()),
	)

	acnt, err := adb.getAccount(addressContainer)
	if err != nil {
		return nil, err
	}
	if acnt == nil {
		return adb.accountFactory.CreateAccount(addressContainer)
	}

	baseAcc, ok := acnt.(baseAccountHandler)
	if ok {
		err = adb.loadCode(baseAcc)
		if err != nil {
			return nil, err
		}

		err = adb.loadDataTrie(baseAcc)
		if err != nil {
			return nil, err
		}
	}

	return acnt, nil
}

func (adb *AccountsDB) getAccount(addressContainer AddressContainer) (AccountHandler, error) {
	addrBytes := addressContainer.Bytes()

	val, err := adb.mainTrie.Get(addrBytes)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil
	}

	acnt, err := adb.accountFactory.CreateAccount(addressContainer)
	if err != nil {
		return nil, err
	}

	err = adb.marshalizer.Unmarshal(acnt, val)
	if err != nil {
		return nil, err
	}

	return acnt, nil
}

// GetExistingAccount returns an existing account if exists or nil if missing
func (adb *AccountsDB) GetExistingAccount(addressContainer AddressContainer) (AccountHandler, error) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	if check.IfNil(addressContainer) {
		return nil, fmt.Errorf("%w in GetExistingAccount, address: %s", ErrNilAddressContainer, hex.EncodeToString(addressContainer.Bytes()))
	}

	log.Trace("accountsDB.GetExistingAccount",
		"address", hex.EncodeToString(addressContainer.Bytes()),
	)

	acnt, err := adb.getAccount(addressContainer)
	if err != nil {
		return nil, err
	}
	if acnt == nil {
		return nil, ErrAccNotFound
	}

	baseAcc, ok := acnt.(baseAccountHandler)
	if ok {
		err = adb.loadCode(baseAcc)
		if err != nil {
			return nil, err
		}

		err = adb.loadDataTrie(baseAcc)
		if err != nil {
			return nil, err
		}
	}

	return acnt, nil
}

// loadCode retrieves and saves the SC code inside AccountState object. Errors if something went wrong
func (adb *AccountsDB) loadCode(accountHandler baseAccountHandler) error {
	if len(accountHandler.GetCodeHash()) == 0 {
		return nil
	}

	val, err := adb.mainTrie.Get(accountHandler.GetCodeHash())
	if err != nil {
		return err
	}

	accountHandler.SetCode(val)
	return nil
}

// RevertToSnapshot apply Revert method over accounts object and removes entries from the list
// Calling with 0 will revert everything. If the snapshot value is out of bounds, an err will be returned
func (adb *AccountsDB) RevertToSnapshot(snapshot int) error {
	log.Trace("accountsDB.RevertToSnapshot started",
		"snapshot", snapshot,
	)

	adb.mutOp.Lock()

	defer func() {
		log.Trace("accountsDB.RevertToSnapshot ended")
		adb.mutOp.Unlock()
	}()

	if snapshot > len(adb.entries) || snapshot < 0 {
		return ErrSnapshotValueOutOfBounds
	}

	for i := len(adb.entries) - 1; i >= snapshot; i-- {
		account, err := adb.entries[i].Revert()
		if err != nil {
			return err
		}

		if !check.IfNil(account) {
			err = adb.saveAccountToTrie(account)
			if err != nil {
				return err
			}
		}
	}

	adb.entries = adb.entries[:snapshot]

	return nil
}

// JournalLen will return the number of entries
func (adb *AccountsDB) JournalLen() int {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	length := len(adb.entries)
	log.Trace("accountsDB.JournalLen",
		"length", length,
	)

	return length
}

// Commit will persist all data inside the trie
func (adb *AccountsDB) Commit() ([]byte, error) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	log.Trace("accountsDB.Commit started")
	adb.entries = make([]JournalEntry, 0)

	oldHashes := make([][]byte, 0)
	//Step 1. commit all data tries
	dataTries := adb.dataTries.GetAll()
	for i := 0; i < len(dataTries); i++ {
		oldTrieHashes := dataTries[i].ResetOldHashes()
		err := dataTries[i].Commit()
		if err != nil {
			return nil, err
		}

		oldHashes = append(oldHashes, oldTrieHashes...)
	}
	adb.dataTries.Reset()

	//Step 2. commit main trie
	adb.mainTrie.AppendToOldHashes(oldHashes)
	err := adb.mainTrie.Commit()
	if err != nil {
		return nil, err
	}

	root, err := adb.mainTrie.Root()
	if err != nil {
		log.Trace("accountsDB.Commit ended", "error", err.Error())
		return nil, err
	}

	log.Trace("accountsDB.Commit ended", "root hash", root)

	return root, nil
}

// RootHash returns the main trie's root hash
func (adb *AccountsDB) RootHash() ([]byte, error) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	rootHash, err := adb.mainTrie.Root()
	log.Trace("accountsDB.RootHash",
		"root hash", rootHash,
		"err", err,
	)

	return rootHash, err
}

// RecreateTrie is used to reload the trie based on an existing rootHash
func (adb *AccountsDB) RecreateTrie(rootHash []byte) error {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	log.Trace("accountsDB.RecreateTrie", "root hash", rootHash)
	defer func() {
		log.Trace("accountsDB.RecreateTrie ended")
	}()

	newTrie, err := adb.mainTrie.Recreate(rootHash)
	if err != nil {
		return err
	}
	if newTrie == nil {
		return ErrNilTrie
	}

	adb.mainTrie = newTrie
	return nil
}

// Journalize adds a new object to entries list.
func (adb *AccountsDB) journalize(entry JournalEntry) {
	if check.IfNil(entry) {
		return
	}

	adb.entries = append(adb.entries, entry)
	log.Trace("accountsDB.Journalize", "new length", len(adb.entries))
}

// PruneTrie removes old values from the trie database
func (adb *AccountsDB) PruneTrie(rootHash []byte, identifier data.TriePruningIdentifier) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	log.Trace("accountsDB.PruneTrie", "root hash", rootHash)

	adb.mainTrie.Prune(rootHash, identifier)
}

// CancelPrune clears the trie's evictionWaitingList
func (adb *AccountsDB) CancelPrune(rootHash []byte, identifier data.TriePruningIdentifier) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	log.Trace("accountsDB.CancelPrune", "root hash", rootHash)

	adb.mainTrie.CancelPrune(rootHash, identifier)
}

// SnapshotState triggers the snapshotting process of the state trie
func (adb *AccountsDB) SnapshotState(rootHash []byte) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	log.Trace("accountsDB.SnapshotState", "root hash", rootHash)

	adb.mainTrie.TakeSnapshot(rootHash)
}

// SetStateCheckpoint sets a checkpoint for the state trie
func (adb *AccountsDB) SetStateCheckpoint(rootHash []byte) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	log.Trace("accountsDB.SetStateCheckpoint", "root hash", rootHash)

	adb.mainTrie.SetCheckpoint(rootHash)
}

// IsPruningEnabled returns true if state pruning is enabled
func (adb *AccountsDB) IsPruningEnabled() bool {
	return adb.mainTrie.IsPruningEnabled()
}

// GetAllLeaves returns all the leaves from a given rootHash
func (adb *AccountsDB) GetAllLeaves(rootHash []byte) (map[string][]byte, error) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	newTrie, err := adb.mainTrie.Recreate(rootHash)
	if err != nil {
		return nil, err
	}
	if newTrie == nil {
		return nil, ErrNilTrie
	}

	allAccounts, err := newTrie.GetAllLeaves()
	if err != nil {
		return nil, err
	}

	return allAccounts, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (adb *AccountsDB) IsInterfaceNil() bool {
	return adb == nil
}

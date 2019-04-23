package state

import (
	"bytes"
	"errors"
	"strconv"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

// AccountsDB is the struct used for accessing accounts
type AccountsDB struct {
	mainTrie       trie.PatriciaMerkelTree
	hasher         hashing.Hasher
	marshalizer    marshal.Marshalizer
	accountFactory AccountFactory

	entries    []JournalEntry
	mutEntries sync.RWMutex
}

// NewAccountsDB creates a new account manager
func NewAccountsDB(
	trie trie.PatriciaMerkelTree,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	accountFactory AccountFactory,
) (*AccountsDB, error) {
	if trie == nil {
		return nil, ErrNilTrie
	}
	if hasher == nil {
		return nil, ErrNilHasher
	}
	if marshalizer == nil {
		return nil, ErrNilMarshalizer
	}
	if accountFactory == nil {
		return nil, ErrNilAccountFactory
	}

	return &AccountsDB{mainTrie: trie,
		hasher:         hasher,
		marshalizer:    marshalizer,
		accountFactory: accountFactory,
		entries:        make([]JournalEntry, 0),
		mutEntries:     sync.RWMutex{},
	}, nil
}

// PutCode sets the SC plain code in AccountWrapper object and trie, code hash in AccountState.
// Errors if something went wrong
func (adb *AccountsDB) PutCode(accountWrapper AccountWrapper, code []byte) error {
	if code == nil || len(code) == 0 {
		return nil
	}
	if accountWrapper == nil {
		return ErrNilAccountWrapper
	}

	codeHash := adb.hasher.Compute(string(code))

	err := adb.addCodeToTrieIfMissing(codeHash, code)
	if err != nil {
		return err
	}

	err = accountWrapper.SetCodeHashWithJournal(codeHash)
	if err != nil {
		return err
	}
	accountWrapper.SetCode(code)

	return nil
}

func (adb *AccountsDB) addCodeToTrieIfMissing(codeHash []byte, code []byte) error {
	val, err := adb.mainTrie.Get(codeHash)
	if err != nil {
		return err
	}
	if val == nil {
		//append a journal entry as the code needs to be inserted in the trie
		entry, err := NewBaseJournalEntryCreation(codeHash, adb.mainTrie)
		if err != nil {
			return err
		}
		adb.Journalize(entry)
		return adb.mainTrie.Update(codeHash, code)
	}

	return nil
}

// RemoveCode deletes the code from the trie. It writes an empty byte slice at codeHash "address"
func (adb *AccountsDB) RemoveCode(codeHash []byte) error {
	return adb.mainTrie.Update(codeHash, make([]byte, 0))
}

// LoadDataTrie retrieves and saves the SC data inside accountWrapper object.
// Errors if something went wrong
func (adb *AccountsDB) LoadDataTrie(accountWrapper AccountWrapper) error {
	if accountWrapper.GetRootHash() == nil {
		//do nothing, the account is either SC library or transfer account
		return nil
	}
	if len(accountWrapper.GetRootHash()) != HashLength {
		return NewErrorTrieNotNormalized(HashLength, len(accountWrapper.GetRootHash()))
	}

	dataTrie, err := adb.mainTrie.Recreate(accountWrapper.GetRootHash(), adb.mainTrie.DBW())
	if err != nil {
		//error as there is an inconsistent state:
		//account has data root hash but does not contain the actual trie
		return NewErrMissingTrie(accountWrapper.GetRootHash())
	}

	accountWrapper.SetDataTrie(dataTrie)
	return nil
}

// SaveData is used to save the data trie (not committing it) and to recompute the new Root value
// If data is not dirtied, method will not create its JournalEntries to keep track of data modification
func (adb *AccountsDB) SaveDataTrie(accountWrapper AccountWrapper) error {
	flagHasDirtyData := false

	if accountWrapper.DataTrie() == nil {
		dataTrie, err := adb.mainTrie.Recreate(make([]byte, 0), adb.mainTrie.DBW())

		if err != nil {
			return err
		}

		accountWrapper.SetDataTrie(dataTrie)
	}

	trackableDataTrie := accountWrapper.DataTrieTracker()
	if trackableDataTrie == nil {
		return ErrNilTrackableDataTrie
	}
	for k, v := range trackableDataTrie.DirtyData() {
		originalValue := trackableDataTrie.OriginalValue([]byte(k))

		if !bytes.Equal(v, originalValue) {
			flagHasDirtyData = true

			err := accountWrapper.DataTrie().Update([]byte(k), v)

			if err != nil {
				return err
			}
		}
	}

	if !flagHasDirtyData {
		//do not need to save, return
		return nil
	}

	//append a journal entry as the data needs to be updated in its trie
	entry, err := NewBaseJournalEntryData(accountWrapper, accountWrapper.DataTrie())
	if err != nil {
		return err
	}

	adb.Journalize(entry)
	err = accountWrapper.SetRootHashWithJournal(accountWrapper.DataTrie().Root())
	if err != nil {
		return err
	}

	trackableDataTrie.ClearDataCaches()

	return adb.SaveAccount(accountWrapper)
}

// HasAccount searches for an account based on the address. Errors if something went wrong and
// outputs if the account exists or not
func (adb *AccountsDB) HasAccount(addressContainer AddressContainer) (bool, error) {
	val, err := adb.mainTrie.Get(addressContainer.Bytes())
	if err != nil {
		return false, err
	}

	return val != nil, nil
}

func (adb *AccountsDB) getAccount(addressContainer AddressContainer) (AccountWrapper, error) {
	addrBytes := addressContainer.Bytes()

	val, err := adb.mainTrie.Get(addrBytes)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil
	}

	acnt := adb.accountFactory.CreateAccount(addressContainer, adb)
	err = adb.marshalizer.Unmarshal(acnt, val)

	return acnt, err
}

// SaveAccount saves the account WITHOUT data trie inside main trie. Errors if something went wrong
func (adb *AccountsDB) SaveAccount(accountWrapper AccountWrapper) error {
	if accountWrapper == nil {
		return errors.New("can not save nil account state")
	}

	//pass the reference to marshalizer, otherwise it will fail marshalizing balance
	buff, err := adb.marshalizer.Marshal(accountWrapper)
	if err != nil {
		return err
	}

	return adb.mainTrie.Update(accountWrapper.AddressContainer().Bytes(), buff)
}

// RemoveAccount removes the account data from underlying trie.
// It basically calls Update with empty slice
func (adb *AccountsDB) RemoveAccount(addressContainer AddressContainer) error {
	return adb.mainTrie.Update(addressContainer.Bytes(), make([]byte, 0))
}

// GetAccountWithJournal fetches the account based on the address. Creates an empty account if the account is missing.
func (adb *AccountsDB) GetAccountWithJournal(addressContainer AddressContainer) (AccountWrapper, error) {
	acnt, err := adb.getAccount(addressContainer)
	if err != nil {
		return nil, err
	}
	if acnt != nil {
		return adb.loadAccountWrapper(acnt, addressContainer)
	}

	return adb.newAccountWrapper(addressContainer)
}

// GetExistingAccount returns an existing account if exists or nil if missing
func (adb *AccountsDB) GetExistingAccount(addressContainer AddressContainer) (AccountWrapper, error) {
	acnt, err := adb.getAccount(addressContainer)
	if err != nil {
		return nil, err
	}
	if acnt == nil {
		return nil, nil
	}

	err = adb.loadCodeAndDataIntoAccountWrapper(acnt)
	if err != nil {
		return nil, err
	}

	return acnt, nil
}

func (adb *AccountsDB) loadAccountWrapper(accountWrapper AccountWrapper, addressContainer AddressContainer) (AccountWrapper, error) {
	err := adb.loadCodeAndDataIntoAccountWrapper(accountWrapper)
	if err != nil {
		return nil, err
	}

	return accountWrapper, nil
}

func (adb *AccountsDB) loadCodeAndDataIntoAccountWrapper(accountWrapper AccountWrapper) error {
	err := adb.loadCode(accountWrapper)
	if err != nil {
		return err
	}

	err = adb.LoadDataTrie(accountWrapper)
	if err != nil {
		return err
	}

	return nil
}

func (adb *AccountsDB) newAccountWrapper(address AddressContainer) (AccountWrapper, error) {
	acnt := adb.accountFactory.CreateAccount(address, adb)
	entry, err := NewBaseJournalEntryCreation(address.Bytes(), adb.mainTrie)
	if err != nil {
		return nil, err
	}

	adb.Journalize(entry)
	err = adb.SaveAccount(acnt)
	if err != nil {
		return nil, err
	}

	return acnt, nil
}

// RevertToSnapshot apply Revert method over accounts object and removes entries from the list
// If snapshot > len(entries) will do nothing, return will be nil
// 0 index based. Calling this method with negative value will do nothing. Calling with 0 revert everything.
// Concurrent safe.
func (adb *AccountsDB) RevertToSnapshot(snapshot int) error {
	if snapshot > len(adb.entries) || snapshot < 0 {
		//outside of bounds array, not quite error, just return
		return nil
	}

	adb.mutEntries.Lock()
	defer adb.mutEntries.Unlock()

	for i := len(adb.entries) - 1; i >= snapshot; i-- {
		account, err := adb.entries[i].Revert()
		if err != nil {
			return err
		}

		_, ok := adb.entries[i].(*BaseJournalEntryRootHash)
		if ok {
			err := adb.LoadDataTrie(account)
			if err != nil {
				return err
			}
		}
		if account != nil {
			err = adb.SaveAccount(account)
		}
	}

	adb.entries = adb.entries[:snapshot]

	return nil
}

// JournalLen will return the number of entries
// Concurrent safe.
func (adb *AccountsDB) JournalLen() int {
	adb.mutEntries.RLock()
	length := len(adb.entries)
	adb.mutEntries.RUnlock()

	return length
}

// Commit will persist all data inside the trie
func (adb *AccountsDB) Commit() ([]byte, error) {
	adb.mutEntries.RLock()
	jEntries := make([]JournalEntry, len(adb.entries))
	copy(jEntries, adb.entries)
	adb.mutEntries.RUnlock()

	//Step 1. iterate through journal entries and commit the data tries accordingly
	for i := 0; i < len(jEntries); i++ {
		jed, found := jEntries[i].(*BaseJournalEntryData)

		if found {
			_, err := jed.Trie().Commit(nil)
			if err != nil {
				return nil, err
			}
		}
	}

	//step 2. clean the journal
	adb.clearJournal()

	//Step 3. commit main trie
	hash, err := adb.mainTrie.Commit(nil)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

// loadCode retrieves and saves the SC code inside AccountState object. Errors if something went wrong
func (adb *AccountsDB) loadCode(accountWrapper AccountWrapper) error {
	if accountWrapper.GetCodeHash() == nil || len(accountWrapper.GetCodeHash()) == 0 {
		return nil
	}
	if len(accountWrapper.GetCodeHash()) != HashLength {
		return errors.New("attempt to search a hash not normalized to" +
			strconv.Itoa(HashLength) + "bytes")
	}

	val, err := adb.mainTrie.Get(accountWrapper.GetCodeHash())
	if err != nil {
		return err
	}

	accountWrapper.SetCode(val)
	return nil
}

// RootHash returns the main trie's root hash
func (adb *AccountsDB) RootHash() []byte {
	return adb.mainTrie.Root()
}

// RecreateTrie is used to reload the trie based on an existing rootHash
func (adb *AccountsDB) RecreateTrie(rootHash []byte) error {
	newTrie, err := adb.mainTrie.Recreate(rootHash, adb.mainTrie.DBW())
	if err != nil {
		return err
	}
	if newTrie == nil {
		return ErrNilTrie
	}

	adb.mainTrie = newTrie
	return nil
}

// AddEntry adds a new object to entries list.
// Concurrent safe.
func (adb *AccountsDB) Journalize(entry JournalEntry) {
	if entry == nil {
		return
	}

	adb.mutEntries.Lock()
	adb.entries = append(adb.entries, entry)
	adb.mutEntries.Unlock()
}

// Clear clears the data from this journal.
func (adb *AccountsDB) clearJournal() {
	adb.mutEntries.Lock()
	adb.entries = make([]JournalEntry, 0)
	adb.mutEntries.Unlock()
}

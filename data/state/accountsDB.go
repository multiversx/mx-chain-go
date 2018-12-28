package state

import (
	"bytes"
	"errors"
	"strconv"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

// AccountsDB is the struct used for accessing accounts
type AccountsDB struct {
	mainTrie    trie.PatriciaMerkelTree
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
	journal     *Journal
}

// NewAccountsDB creates a new account manager
func NewAccountsDB(trie trie.PatriciaMerkelTree, hasher hashing.Hasher, marshalizer marshal.Marshalizer) (*AccountsDB, error) {
	if trie == nil {
		return nil, ErrNilTrie
	}

	if hasher == nil {
		return nil, ErrNilHasher
	}

	if marshalizer == nil {
		return nil, ErrNilMarshalizer
	}

	return &AccountsDB{mainTrie: trie, hasher: hasher, marshalizer: marshalizer, journal: NewJournal()}, nil
}

// PutCode sets the SC plain code in journalizedAccountWrapper object and trie, code hash in AccountState.
// Errors if something went wrong
func (adb *AccountsDB) PutCode(journalizedAccountWrapper JournalizedAccountWrapper, code []byte) error {
	if code == nil || len(code) == 0 {
		return nil
	}

	if adb.mainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	if journalizedAccountWrapper == nil {
		return ErrNilJurnalizingAccountWrapper
	}

	codeHash := adb.hasher.Compute(string(code))

	err := adb.addCodeToTrieIfMissing(codeHash, code)
	if err != nil {
		return err
	}

	err = journalizedAccountWrapper.SetCodeHashWithJournal(codeHash)
	if err != nil {
		return err
	}
	journalizedAccountWrapper.SetCode(code)

	return nil
}

func (adb *AccountsDB) addCodeToTrieIfMissing(codeHash []byte, code []byte) error {
	val, err := adb.mainTrie.Get(codeHash)
	if err != nil {
		return err
	}

	if val == nil {
		//append a journal entry as the code needs to be inserted in the trie
		adb.AddJournalEntry(NewJournalEntryCode(codeHash))
		return adb.mainTrie.Update(codeHash, code)
	}

	return nil
}

// RemoveCode deletes the code from the trie. It writes an empty byte slice at codeHash "address"
func (adb *AccountsDB) RemoveCode(codeHash []byte) error {
	if adb.mainTrie == nil {
		return ErrNilTrie
	}

	return adb.mainTrie.Update(codeHash, make([]byte, 0))
}

// LoadDataTrie retrieves and saves the SC data inside accountWrapper object.
// Errors if something went wrong
func (adb *AccountsDB) LoadDataTrie(accountWrapper AccountWrapper) error {
	if accountWrapper.BaseAccount().RootHash == nil {
		//do nothing, the account is either SC library or transfer account
		return nil
	}

	if adb.mainTrie == nil {
		return ErrNilTrie
	}

	if len(accountWrapper.BaseAccount().RootHash) != HashLength {
		return NewErrorTrieNotNormalized(HashLength, len(accountWrapper.BaseAccount().RootHash))
	}

	dataTrie, err := adb.mainTrie.Recreate(accountWrapper.BaseAccount().RootHash, adb.mainTrie.DBW())
	if err != nil {
		//error as there is an inconsistent state:
		//account has data root hash but does not contain the actual trie
		return NewErrMissingTrie(accountWrapper.BaseAccount().RootHash)
	}

	accountWrapper.SetDataTrie(dataTrie)
	return nil
}

// SaveData is used to save the data trie (not committing it) and to recompute the new Root value
// If data is not dirtied, method will not create its JournalEntries to keep track of data modification
func (adb *AccountsDB) SaveData(journalizedAccountWrapper JournalizedAccountWrapper) error {
	if adb.mainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	flagHasDirtyData := false

	if journalizedAccountWrapper.DataTrie() == nil {
		dataTrie, err := adb.mainTrie.Recreate(make([]byte, 0), adb.mainTrie.DBW())

		if err != nil {
			return err
		}

		journalizedAccountWrapper.SetDataTrie(dataTrie)
	}

	for k, v := range journalizedAccountWrapper.DirtyData() {
		originalValue := journalizedAccountWrapper.OriginalValue([]byte(k))

		if !bytes.Equal(v, originalValue) {
			flagHasDirtyData = true

			err := journalizedAccountWrapper.DataTrie().Update([]byte(k), v)

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
	adb.AddJournalEntry(NewJournalEntryData(journalizedAccountWrapper, journalizedAccountWrapper.DataTrie()))
	err := journalizedAccountWrapper.SetRootHashWithJournal(journalizedAccountWrapper.DataTrie().Root())
	if err != nil {
		return err
	}
	journalizedAccountWrapper.ClearDataCaches()
	return adb.SaveJournalizedAccount(journalizedAccountWrapper)
}

func (adb *AccountsDB) getAccount(addressContainer AddressContainer) (*Account, error) {
	if adb.mainTrie == nil {
		return nil, errors.New("attempt to search on a nil trie")
	}

	addressHash := adb.hasher.Compute(string(addressContainer.Bytes()))
	val, err := adb.mainTrie.Get(addressHash)
	if err != nil {
		return nil, err
	}

	if val == nil {
		return nil, nil
	}

	acnt := NewAccount()

	err = adb.marshalizer.Unmarshal(acnt, val)
	if err != nil {
		return nil, err
	}

	return acnt, err
}

// HasAccountState searches for an account based on the address. Errors if something went wrong and
// outputs if the account exists or not
func (adb *AccountsDB) HasAccount(addressContainer AddressContainer) (bool, error) {
	if adb.mainTrie == nil {
		return false, errors.New("attempt to search on a nil trie")
	}

	addressHash := adb.hasher.Compute(string(addressContainer.Bytes()))
	val, err := adb.mainTrie.Get(addressHash)

	if err != nil {
		return false, err
	}

	return val != nil, nil
}

// SaveJournalizedAccount saves the account WITHOUT data trie inside main trie. Errors if something went wrong
func (adb *AccountsDB) SaveJournalizedAccount(journalizedAccountWrapper JournalizedAccountWrapper) error {
	if journalizedAccountWrapper == nil {
		return errors.New("can not save nil account state")
	}

	if adb.mainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	//create an Account object that will be serialized
	account := journalizedAccountWrapper.BaseAccount()

	//pass the reference to marshalizer, otherwise it will fail marshalizing balance
	buff, err := adb.marshalizer.Marshal(account)

	if err != nil {
		return err
	}

	addressHash := adb.hasher.Compute(string(journalizedAccountWrapper.AddressContainer().Bytes()))
	return adb.mainTrie.Update(addressHash, buff)
}

// RemoveAccount removes the account data from underlying trie.
// It basically calls Update with empty slice
func (adb *AccountsDB) RemoveAccount(addressContainer AddressContainer) error {
	if adb.mainTrie == nil {
		return ErrNilTrie
	}

	addressHash := adb.hasher.Compute(string(addressContainer.Bytes()))
	return adb.mainTrie.Update(addressHash, make([]byte, 0))
}

// GetJournalizedAccount fetches the account based on the address. Creates an empty account if the account is missing.
func (adb *AccountsDB) GetJournalizedAccount(addressContainer AddressContainer) (JournalizedAccountWrapper, error) {
	acnt, err := adb.getAccount(addressContainer)

	if err != nil {
		return nil, err
	}

	if acnt != nil {
		return adb.loadJournalizedAccountWrapper(acnt, addressContainer)
	}

	return adb.newJournalizedAccountWrapper(addressContainer)
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

	accountWrap, err := NewAccountWrap(addressContainer, acnt)
	if err != nil {
		return nil, err
	}

	err = adb.loadCodeAndDataIntoAccountWrapper(accountWrap)

	if err != nil {
		return nil, err
	}

	return accountWrap, nil
}

func (adb *AccountsDB) loadJournalizedAccountWrapper(account *Account, addressContainer AddressContainer) (JournalizedAccountWrapper, error) {
	journalizedAccountWrap, err := NewJournalizedAccountWrapFromAccountContainer(addressContainer, account, adb)
	if err != nil {
		return nil, err
	}

	err = adb.loadCodeAndDataIntoAccountWrapper(journalizedAccountWrap)
	if err != nil {
		return nil, err
	}

	return journalizedAccountWrap, nil
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

func (adb *AccountsDB) newJournalizedAccountWrapper(addressHandler AddressContainer) (JournalizedAccountWrapper, error) {
	account := NewAccount()

	journalizedAccountWrap, err := NewJournalizedAccountWrapFromAccountContainer(addressHandler, account, adb)
	if err != nil {
		return nil, err
	}

	adb.journal.AddEntry(NewJournalEntryCreation(addressHandler))

	err = adb.SaveJournalizedAccount(journalizedAccountWrap)
	if err != nil {
		return nil, err
	}

	return journalizedAccountWrap, nil
}

// AddJournalEntry adds a new object to entries list.
// Concurrent safe.
func (adb *AccountsDB) AddJournalEntry(je JournalEntry) {
	adb.journal.AddEntry(je)
}

// RevertToSnapshot apply Revert method over accounts object and removes entries from the list
// If snapshot > len(entries) will do nothing, return will be nil
// 0 index based. Calling this method with negative value will do nothing. Calling with 0 revert everything.
// Concurrent safe.
func (adb *AccountsDB) RevertToSnapshot(snapshot int) error {
	return adb.journal.RevertToSnapshot(snapshot, adb)
}

// JournalLen will return the number of entries
// Concurrent safe.
func (adb *AccountsDB) JournalLen() int {
	return adb.journal.Len()
}

// Commit will persist all data inside the trie
func (adb *AccountsDB) Commit() ([]byte, error) {
	jEntries := adb.journal.Entries()

	//Step 1. iterate through journal entries and commit the data tries accordingly
	for i := 0; i < len(jEntries); i++ {
		jed, found := jEntries[i].(*JournalEntryData)

		if found {
			_, err := jed.Trie().Commit(nil)
			if err != nil {
				return nil, err
			}
		}
	}

	//step 2. clean the journal
	adb.journal.Clear()

	//Step 3. commit main trie
	hash, err := adb.mainTrie.Commit(nil)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

// loadCode retrieves and saves the SC code inside AccountState object. Errors if something went wrong
func (adb *AccountsDB) loadCode(accountWrapper AccountWrapper) error {
	if adb.mainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	if accountWrapper.BaseAccount().CodeHash == nil || len(accountWrapper.BaseAccount().CodeHash) == 0 {
		return nil
	}

	if len(accountWrapper.BaseAccount().CodeHash) != HashLength {
		return errors.New("attempt to search a hash not normalized to" +
			strconv.Itoa(HashLength) + "bytes")
	}

	val, err := adb.mainTrie.Get(accountWrapper.BaseAccount().CodeHash)

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

	adb.mainTrie = newTrie
	return nil
}

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
	MainTrie trie.PatriciaMerkelTree
	hasher   hashing.Hasher
	marsh    marshal.Marshalizer

	journal *Journal
}

// NewAccountsDB creates a new account manager
func NewAccountsDB(trie trie.PatriciaMerkelTree, hasher hashing.Hasher, marsh marshal.Marshalizer) *AccountsDB {
	adb := AccountsDB{MainTrie: trie, hasher: hasher, marsh: marsh}
	adb.journal = NewJournal()

	return &adb
}

// PutCode sets the SC plain code in AccountState object and trie, code hash in AccountState.
// Errors if something went wrong
func (adb *AccountsDB) PutCode(accountStateHandler AccountStateHandler, code []byte) error {
	if (code == nil) || (len(code) == 0) {
		return nil
	}

	if adb.MainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	codeHash := adb.hasher.Compute(string(code))

	if accountStateHandler != nil {
		err := accountStateHandler.SetCodeHash(adb, codeHash)
		if err != nil {
			return err
		}
		accountStateHandler.SetCode(code)
	}

	//test that we add the code, or reuse an existing code as to be able to correctly revert the addition
	val, err := adb.MainTrie.Get(codeHash)
	if err != nil {
		return err
	}

	if val == nil {
		//append a journal entry as the code needs to be inserted in the trie
		adb.AddJournalEntry(NewJournalEntryCode(codeHash))
		return adb.MainTrie.Update(codeHash, code)
	}

	return nil
}

// RemoveCode deletes the code from the trie. It writes an empty byte slice at codeHash "address"
func (adb *AccountsDB) RemoveCode(codeHash []byte) error {
	return adb.MainTrie.Update(codeHash, make([]byte, 0))
}

// RetrieveDataTrie retrieves and saves the SC data inside AccountState object. Errors if something went wrong
func (adb *AccountsDB) RetrieveDataTrie(accountStateHandler AccountStateHandler) error {
	if accountStateHandler.RootHash() == nil {
		//do nothing, the account is either SC library or transfer account
		return nil
	}

	if adb.MainTrie == nil {
		return ErrNilTrie
	}

	if len(accountStateHandler.RootHash()) != HashLength {
		return NewErrorTrieNotNormalized(HashLength, len(accountStateHandler.RootHash()))
	}

	dataTrie, err := adb.MainTrie.Recreate(accountStateHandler.RootHash(), adb.MainTrie.DBW())
	if err != nil {
		//error as there is an inconsistent state:
		//account has data root hash but does not contain the actual trie
		return NewErrMissingTrie(accountStateHandler.RootHash())
	}

	accountStateHandler.SetDataTrie(dataTrie)
	return nil
}

// SaveData is used to save the data trie (not committing it) and to recompute the new Root value
// If data is not dirtied, method will not create its JournalEntries to keep track of data modification
func (adb *AccountsDB) SaveData(accountStateHandler AccountStateHandler) error {
	if adb.MainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	flagHasDirtyData := false

	if accountStateHandler.DataTrie() == nil {
		dataTrie, err := adb.MainTrie.Recreate(make([]byte, 0), adb.MainTrie.DBW())

		if err != nil {
			return err
		}

		accountStateHandler.SetDataTrie(dataTrie)
	}

	for k, v := range accountStateHandler.DirtyData() {
		originalValue := accountStateHandler.OriginalValue([]byte(k))

		if !bytes.Equal(v, originalValue) {
			flagHasDirtyData = true

			err := accountStateHandler.DataTrie().Update([]byte(k), v)

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
	adb.AddJournalEntry(NewJournalEntryData(accountStateHandler.DataTrie(), accountStateHandler))
	err := accountStateHandler.SetRootHash(adb, accountStateHandler.DataTrie().Root())
	if err != nil {
		return err
	}
	accountStateHandler.ClearDataCaches()
	return adb.SaveAccountState(accountStateHandler)
}

// HasAccountState searches for an account based on the address. Errors if something went wrong and
// outputs if the account exists or not
func (adb *AccountsDB) getAccount(addressHandler AddressHandler) (*Account, error) {
	if adb.MainTrie == nil {
		return nil, errors.New("attempt to search on a nil trie")
	}

	val, err := adb.MainTrie.Get(addressHandler.Hash(adb.hasher))
	if err != nil {
		return nil, err
	}

	if val == nil {
		return nil, nil
	}

	acnt := NewAccount()

	err = adb.marsh.Unmarshal(acnt, val)
	if err != nil {
		return nil, err
	}

	return acnt, err
}

// HasAccountState searches for an account based on the address. Errors if something went wrong and
// outputs if the account exists or not
func (adb *AccountsDB) HasAccountState(addressHandler AddressHandler) (bool, error) {
	if adb.MainTrie == nil {
		return false, errors.New("attempt to search on a nil trie")
	}

	val, err := adb.MainTrie.Get(addressHandler.Hash(adb.hasher))

	if err != nil {
		return false, err
	}

	return val != nil, nil
}

// SaveAccountState saves the account WITHOUT data trie inside main trie. Errors if something went wrong
func (adb *AccountsDB) SaveAccountState(accountStateHandler AccountStateHandler) error {
	if accountStateHandler == nil {
		return errors.New("can not save nil account state")
	}

	if adb.MainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	//create Account that will be serialized
	balance := accountStateHandler.Balance()
	acnt := Account{
		Nonce:    accountStateHandler.Nonce(),
		Balance:  &balance,
		CodeHash: accountStateHandler.CodeHash(),
		RootHash: accountStateHandler.RootHash(),
	}

	buff, err := adb.marsh.Marshal(acnt)

	if err != nil {
		return err
	}

	return adb.MainTrie.Update(accountStateHandler.AddressHandler().Hash(adb.hasher), buff)
}

// RemoveAccount removes the account data from underlying trie.
// It basically calls Update with empty slice
func (adb *AccountsDB) RemoveAccount(addressHandler AddressHandler) error {
	if adb.MainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	return adb.MainTrie.Update(addressHandler.Hash(adb.hasher), make([]byte, 0))
}

// GetOrCreateAccountState fetches the account based on the address. Creates an empty account if the account is missing.
func (adb *AccountsDB) GetOrCreateAccountState(addressHandler AddressHandler) (AccountStateHandler, error) {
	acnt, err := adb.getAccount(addressHandler)

	if err != nil {
		return nil, err
	}

	if acnt != nil {
		return adb.loadAccountState(acnt, addressHandler)
	}

	return adb.newAccountState(addressHandler)
}

func (adb *AccountsDB) loadAccountState(account *Account, addressHandler AddressHandler) (AccountStateHandler, error) {
	state := NewAccountState(addressHandler, account, adb.hasher)
	err := adb.retrieveCode(state)

	if err != nil {
		return nil, err
	}

	err = adb.RetrieveDataTrie(state)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (adb *AccountsDB) newAccountState(addressHandler AddressHandler) (AccountStateHandler, error) {
	acState := NewAccountState(addressHandler, NewAccount(), adb.hasher)
	adb.journal.AddEntry(NewJournalEntryCreation(addressHandler))

	err := adb.SaveAccountState(acState)
	if err != nil {
		return nil, err
	}

	return acState, nil
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
	hash, err := adb.MainTrie.Commit(nil)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

// retrieveCode retrieves and saves the SC code inside AccountState object. Errors if something went wrong
func (adb *AccountsDB) retrieveCode(accountStateHandler AccountStateHandler) error {
	if accountStateHandler.CodeHash() == nil {
		accountStateHandler.SetCode(nil)
		return nil
	}

	if adb.MainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	if len(accountStateHandler.CodeHash()) != HashLength {
		return errors.New("attempt to search a hash not normalized to" +
			strconv.Itoa(HashLength) + "bytes")
	}

	val, err := adb.MainTrie.Get(accountStateHandler.CodeHash())

	if err != nil {
		return err
	}

	//TODO implement some kind of checking of code against its codehash

	accountStateHandler.SetCode(val)
	return nil
}

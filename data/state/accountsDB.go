package state

import (
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
func NewAccountsDB(tr trie.PatriciaMerkelTree, hasher hashing.Hasher, marsh marshal.Marshalizer) *AccountsDB {
	adb := AccountsDB{MainTrie: tr, hasher: hasher, marsh: marsh}
	adb.journal = NewJournal()

	return &adb
}

// PutCode sets the SC plain code in AccountState object and trie, code hash in AccountState.
// Errors if something went wrong
func (adb *AccountsDB) PutCode(acState AccountStateHandler, code []byte) error {
	if (code == nil) || (len(code) == 0) {
		return nil
	}

	if adb.MainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	codeHash := adb.hasher.Compute(string(code))

	if acState != nil {
		err := acState.SetCodeHash(adb, codeHash)
		if err != nil {
			return err
		}
		acState.SetCode(code)
	}

	//test that we add the code, or reuse an existing code as to be able to correctly revert the addition
	val, err := adb.MainTrie.Get(codeHash)
	if err != nil {
		return err
	}

	if val == nil {
		//append a journal entry as the code needs to be inserted in the trie
		adb.AddJurnalEntry(NewJournalEntryCode(codeHash))
		return adb.MainTrie.Update(codeHash, code)
	}

	return nil
}

// RemoveCode deletes the code from the trie. It writes an empty byte slice at codeHash "address"
func (adb *AccountsDB) RemoveCode(codeHash []byte) error {
	return adb.MainTrie.Update(codeHash, make([]byte, 0))
}

// RetrieveDataTrie retrieves and saves the SC data inside AccountState object. Errors if something went wrong
func (adb *AccountsDB) RetrieveDataTrie(acState AccountStateHandler) error {
	if acState.Root() == nil {
		//do nothing, the account is either SC library or transfer account
		return nil
	}

	if adb.MainTrie == nil {
		return ErrNilTrie
	}

	if len(acState.Root()) != HashLength {
		return NewErrorTrieNotNormalized(HashLength, len(acState.Root()))
	}

	dataTrie, err := adb.MainTrie.Recreate(acState.Root(), adb.MainTrie.DBW())
	if err != nil {
		//error as there is an inconsistent state:
		//account has data root but does not contain the actual trie
		return NewErrMissingTrie(acState.Root())
	}

	acState.SetDataTrie(dataTrie)
	return nil
}

// SaveData is used to save the data trie (not committing it) and to recompute the new Root value
// If data is not dirtied, method will not create its JournalEntries to keep track of data modification
func (adb *AccountsDB) SaveData(acState AccountStateHandler) error {
	if adb.MainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	//collapse dirty as to check if we really need to save some data
	acState.CollapseDirty()

	if len(acState.DirtyData()) == 0 {
		//do not need to save, return
		return nil
	}

	if acState.DataTrie() == nil {
		dataTrie, err := adb.MainTrie.Recreate(make([]byte, 0), adb.MainTrie.DBW())

		if err != nil {
			return err
		}

		acState.SetDataTrie(dataTrie)
	}

	for k, v := range acState.DirtyData() {
		err := acState.DataTrie().Update([]byte(k), v)

		if err != nil {
			return err
		}
	}

	//append a journal entry as the data needs to be updated in its trie
	adb.AddJurnalEntry(NewJournalEntryData(acState.DataTrie(), acState))
	err := acState.SetRoot(adb, acState.DataTrie().Root())
	if err != nil {
		return err
	}
	return adb.SaveAccountState(acState)
}

// HasAccountState searches for an account based on the address. Errors if something went wrong and
// outputs if the account exists or not
func (adb *AccountsDB) HasAccountState(address AddressHandler) (bool, error) {
	if adb.MainTrie == nil {
		return false, errors.New("attempt to search on a nil trie")
	}

	val, err := adb.MainTrie.Get(address.Hash(adb.hasher))

	if err != nil {
		return false, err
	}

	return val != nil, nil
}

// SaveAccountState saves the account WITHOUT data trie inside main trie. Errors if something went wrong
func (adb *AccountsDB) SaveAccountState(acState AccountStateHandler) error {
	if acState == nil {
		return errors.New("can not save nil account state")
	}

	if adb.MainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	//create Account that will be serialized
	balance := acState.Balance()
	acnt := Account{
		Nonce:    acState.Nonce(),
		Balance:  &balance,
		CodeHash: acState.CodeHash(),
		Root:     acState.Root(),
	}

	buff, err := adb.marsh.Marshal(acnt)

	if err != nil {
		return err
	}

	err = adb.MainTrie.Update(acState.Address().Hash(adb.hasher), buff)
	if err != nil {
		return err
	}

	return nil
}

// RemoveAccount removes the account data from underlying trie.
// It basically calls Update with empty slice
func (adb *AccountsDB) RemoveAccount(address AddressHandler) error {
	if adb.MainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	return adb.MainTrie.Update(address.Hash(adb.hasher), make([]byte, 0))
}

// GetOrCreateAccountState fetches the account based on the address. Creates an empty account if the account is missing.
func (adb *AccountsDB) GetOrCreateAccountState(address AddressHandler) (AccountStateHandler, error) {
	has, err := adb.HasAccountState(address)
	if err != nil {
		return nil, err
	}

	if has {
		val, err := adb.MainTrie.Get(address.Hash(adb.hasher))
		if err != nil {
			return nil, err
		}

		acnt := NewAccount()

		err = adb.marsh.Unmarshal(acnt, val)
		if err != nil {
			return nil, err
		}

		state := NewAccountState(address, acnt, adb.hasher)
		err = adb.retrieveCode(state)
		if err != nil {
			return nil, err
		}
		err = adb.RetrieveDataTrie(state)
		if err != nil {
			return nil, err
		}

		return state, nil
	}

	acState := NewAccountState(address, NewAccount(), adb.hasher)
	adb.journal.AddEntry(NewJournalEntryCreation(address))

	err = adb.SaveAccountState(acState)
	if err != nil {
		return nil, err
	}

	return acState, nil
}

// AddJurnalEntry adds a new object to entries list.
// Concurrent safe.
func (adb *AccountsDB) AddJurnalEntry(je JournalEntry) {
	adb.journal.AddEntry(je)
}

// RevertFromSnapshot apply Revert method over accounts object and removes entries from the list
// If snapshot > len(entries) will do nothing, return will be nil
// 0 index based. Calling this method with negative value will do nothing. Calling with 0 revert everything.
// Concurrent safe.
func (adb *AccountsDB) RevertFromSnapshot(snapshot int) error {
	return adb.journal.RevertFromSnapshot(snapshot, adb)
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
func (adb *AccountsDB) retrieveCode(state AccountStateHandler) error {
	if state.CodeHash() == nil {
		state.SetCode(nil)
		return nil
	}

	if adb.MainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	if len(state.CodeHash()) != HashLength {
		return errors.New("attempt to search a hash not normalized to" +
			strconv.Itoa(HashLength) + "bytes")
	}

	val, err := adb.MainTrie.Get(state.CodeHash())

	if err != nil {
		return err
	}

	//TODO implement some kind of checking of code against its codehash

	state.SetCode(val)
	return nil
}

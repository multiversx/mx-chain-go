package state

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
)

// Account is a struct that will be serialized/deserialized
type Account struct {
	Nonce    uint64
	Balance  *big.Int
	CodeHash []byte
	Root     []byte
}

// NewAccount creates a new account object making sure all fields are properly initialized
func NewAccount() *Account {
	acnt := Account{}
	acnt.Balance = big.NewInt(0)

	return &acnt
}

// AccountState is a struct that wraps Account and add functionalities to it
type AccountState struct {
	account      *Account
	Addr         Address
	Code         []byte
	DataTrie     trie.PatriciaMerkelTree
	hasher       hashing.Hasher
	prevRoot     []byte
	originalData map[string][]byte
	dirtyData    map[string][]byte
}

// NewAccountState creates new wrapper for an Account (that has just been retrieved)
func NewAccountState(address Address, account *Account, hasher hashing.Hasher) *AccountState {
	acState := AccountState{
		account:      account,
		Addr:         address,
		prevRoot:     account.Root,
		originalData: make(map[string][]byte),
		dirtyData:    make(map[string][]byte),
	}

	acState.hasher = hasher

	return &acState
}

// Nonce will return a uint64 that holds the nonce of the account
func (as *AccountState) Nonce() uint64 {
	return as.account.Nonce
}

// SetNonce sets the nonce and also records the JurnalEntry
func (as *AccountState) SetNonce(accounts AccountsHandler, nonce uint64) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	accounts.Jurnal().AddEntry(NewJurnalEntryNonce(&as.Addr, as, as.account.Nonce))
	as.account.Nonce = nonce
	accounts.SaveAccountState(as)
	return nil
}

// Balance will return a copy of big.Int that holds the balance
// Not using pointer as to keep track of all balance changes
func (as *AccountState) Balance() big.Int {
	return *as.account.Balance
}

// SetBalance sets the balance and also records the JurnalEntry
func (as *AccountState) SetBalance(accounts AccountsHandler, value *big.Int) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	if value == nil {
		return ErrNilValue
	}

	accounts.Jurnal().AddEntry(NewJurnalEntryBalance(&as.Addr, as, as.account.Balance))
	as.account.Balance = value
	accounts.SaveAccountState(as)
	return nil
}

// CodeHash returns the account's code hash if it is a SC account. Nil for transfer account
func (as *AccountState) CodeHash() []byte {
	return as.account.CodeHash
}

// SetCodeHash sets the code hash on the current account. The account will be considered as a SC account
func (as *AccountState) SetCodeHash(accounts AccountsHandler, codeHash []byte) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	accounts.Jurnal().AddEntry(NewJurnalEntryCodeHash(&as.Addr, as, as.account.CodeHash))
	as.account.CodeHash = codeHash
	return nil
}

// Root returns the account's (data) root. Nil for library SC
func (as *AccountState) Root() []byte {
	return as.account.Root
}

// SetRoot sets the (data) root on the current account. The account will be considered as a data related SC account
func (as *AccountState) SetRoot(accounts AccountsHandler, root []byte) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	accounts.Jurnal().AddEntry(NewJurnalEntryRoot(&as.Addr, as, as.account.Root))
	as.account.Root = root
	return nil
}

// RetrieveValue fetches the value from a particular key searching the account data store
// The search starts with dirty map, continues with original map and ends with the trie
// Data must have been retrieved from its trie
func (as *AccountState) RetrieveValue(key []byte) ([]byte, error) {
	if as.DataTrie == nil {
		return nil, ErrNilDataTrie
	}

	strKey := string(key)

	//search in dirty data cache
	data, found := as.dirtyData[strKey]
	if found {
		return data, nil
	}

	//search in original data cache
	data, found = as.originalData[strKey]
	if found {
		return data, nil
	}

	//ok, not in cache, retrieve from trie
	data, err := as.DataTrie.Get(key)
	if err != nil {
		return nil, err
	}

	//got the value, put it originalData cache as the next fetch will run faster
	as.originalData[string(key)] = data
	return data, nil
}

// SaveKeyValue stores in dirtyData the data keys "touched"
// It does not care if the data is really dirty as calling this check here will be sub-optimal
func (as *AccountState) SaveKeyValue(key []byte, value []byte) {
	as.dirtyData[string(key)] = value
}

// CollapseDirty filters out the data from dirtyData map that isn't dirty
func (as *AccountState) CollapseDirty() {
	for k, v := range as.dirtyData {
		originalValue := as.originalData[k]

		if bytes.Equal(v, originalValue) {
			//safely deletes the record
			delete(as.dirtyData, k)
		}
	}
}

// ClearDirty empties the dirtyData map
func (as *AccountState) ClearDirty() {
	as.dirtyData = make(map[string][]byte)
}

// DirtyData returns the map of (key, value) pairs that contain the data needed to be saved in the data trie
func (as *AccountState) DirtyData() map[string][]byte {
	return as.dirtyData
}

// Account retrieves the account pointer for accessing raw account data and bypassing JurnalEntries creation
// This pointer will be used for serializing/deserializing account data
func (as *AccountState) Account() *Account {
	return as.account
}

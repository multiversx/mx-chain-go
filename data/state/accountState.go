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
	address      AddressHandler
	code         []byte
	dataTrie     trie.PatriciaMerkelTree
	hasher       hashing.Hasher
	originalData map[string][]byte
	dirtyData    map[string][]byte
}

// NewAccountState creates new wrapper for an Account (that has just been retrieved)
func NewAccountState(address AddressHandler, account *Account, hasher hashing.Hasher) *AccountState {
	acState := AccountState{
		account:      account,
		address:      address,
		originalData: make(map[string][]byte),
		dirtyData:    make(map[string][]byte),
	}

	acState.hasher = hasher

	return &acState
}

// Address returns the current account state's handler
func (as *AccountState) Address() AddressHandler {
	return as.address
}

// Nonce will return a uint64 that holds the nonce of the account
func (as *AccountState) Nonce() uint64 {
	return as.account.Nonce
}

// SetNonce sets the nonce and also records the JournalEntry
func (as *AccountState) SetNonce(accounts AccountsHandler, nonce uint64) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	accounts.AddJurnalEntry(NewJournalEntryNonce(as.address, as, as.account.Nonce))
	as.account.Nonce = nonce
	return accounts.SaveAccountState(as)
}

// SetNonceNoJournal sets the nonce without making a Journal entry
// Should be used only in reverts
func (as *AccountState) SetNonceNoJournal(nonce uint64) {
	as.account.Nonce = nonce
}

// Balance will return a copy of big.Int that holds the balance
// Not using pointer as to keep track of all balance changes
func (as *AccountState) Balance() big.Int {
	return *as.account.Balance
}

// SetBalance sets the balance and also records the JournalEntry
func (as *AccountState) SetBalance(accounts AccountsHandler, value *big.Int) error {
	if accounts == nil {
		return ErrNilAccountsHandler
	}

	if value == nil {
		return ErrNilValue
	}

	accounts.AddJurnalEntry(NewJournalEntryBalance(as.address, as, as.account.Balance))
	as.account.Balance = value
	return accounts.SaveAccountState(as)
}

// SetBalanceNoJournal sets the balance without making a Journal entry
// Should be used only in reverts
func (as *AccountState) SetBalanceNoJournal(value *big.Int) {
	as.account.Balance = value
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

	accounts.AddJurnalEntry(NewJournalEntryCodeHash(as.address, as, as.account.CodeHash))
	as.account.CodeHash = codeHash
	return nil
}

// SetCodeHashNoJournal sets the code hash without making a Journal entry
// Should be used only in reverts
func (as *AccountState) SetCodeHashNoJournal(codeHash []byte) {
	as.account.CodeHash = codeHash
}

// Code gets the actual code that needs to be run in the VM
func (as *AccountState) Code() []byte {
	return as.code
}

// SetCode sets the actual code that needs to be run in the VM
func (as *AccountState) SetCode(code []byte) {
	as.code = code
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

	accounts.AddJurnalEntry(NewJournalEntryRoot(as.address, as, as.account.Root))
	as.account.Root = root
	return nil
}

// SetRootNoJournal sets the root data hash without making a Journal entry
// Should be used only in reverts
func (as *AccountState) SetRootNoJournal(root []byte) {
	as.account.Root = root
}

// DataTrie returns the trie that holds the current account's data
func (as *AccountState) DataTrie() trie.PatriciaMerkelTree {
	return as.dataTrie
}

// SetDataTrie sets the trie that holds the current account's data
func (as *AccountState) SetDataTrie(t trie.PatriciaMerkelTree) {
	as.dataTrie = t
}

// RetrieveValue fetches the value from a particular key searching the account data store
// The search starts with dirty map, continues with original map and ends with the trie
// Data must have been retrieved from its trie
func (as *AccountState) RetrieveValue(key []byte) ([]byte, error) {
	if as.dataTrie == nil {
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
	data, err := as.dataTrie.Get(key)
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

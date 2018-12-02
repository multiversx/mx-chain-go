package state

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
)

// Account is a struct that will be serialized/deserialized inside the trie
// CodeHash and RootHash are hashes (pointers) to the location where the actual data resides
type Account struct {
	Nonce    uint64
	Balance  *big.Int
	CodeHash []byte
	RootHash []byte
}

// NewAccount creates a new account object making sure all fields are properly initialized
func NewAccount() *Account {
	acnt := Account{}
	acnt.Balance = big.NewInt(0)

	return &acnt
}

// AccountState is a wrapper over Account that holds the actual data
// (Code is the actual byte-code array that will need to be executed in VM and
// DataTrie holds the trie that contains the variable data)
type AccountState struct {
	account        *Account
	addressHandler AddressHandler
	code           []byte
	dataTrie       trie.PatriciaMerkelTree
	hasher         hashing.Hasher
	originalData   map[string][]byte
	dirtyData      map[string][]byte
}

// NewAccountState creates new wrapper for an Account (that has just been retrieved)
func NewAccountState(addressHandler AddressHandler, account *Account, hasher hashing.Hasher) *AccountState {
	acState := AccountState{
		account:        account,
		addressHandler: addressHandler,
		originalData:   make(map[string][]byte),
		dirtyData:      make(map[string][]byte),
	}

	acState.hasher = hasher

	return &acState
}

// AddressHandler returns the current account state's handler
func (as *AccountState) AddressHandler() AddressHandler {
	return as.addressHandler
}

// Nonce will return a uint64 that holds the nonce of the account
func (as *AccountState) Nonce() uint64 {
	return as.account.Nonce
}

// SetNonce sets the nonce and also records the JournalEntry
func (as *AccountState) SetNonce(accountsHandler AccountsHandler, nonce uint64) error {
	if accountsHandler == nil {
		return ErrNilAccountsHandler
	}

	accountsHandler.AddJournalEntry(NewJournalEntryNonce(as.addressHandler, as, as.account.Nonce))
	as.account.Nonce = nonce
	return accountsHandler.SaveAccountState(as)
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
func (as *AccountState) SetBalance(accountsHandler AccountsHandler, value *big.Int) error {
	if accountsHandler == nil {
		return ErrNilAccountsHandler
	}

	if value == nil {
		return ErrNilValue
	}

	accountsHandler.AddJournalEntry(NewJournalEntryBalance(as.addressHandler, as, as.account.Balance))
	as.account.Balance = value
	return accountsHandler.SaveAccountState(as)
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
func (as *AccountState) SetCodeHash(accountsHandler AccountsHandler, codeHash []byte) error {
	if accountsHandler == nil {
		return ErrNilAccountsHandler
	}

	accountsHandler.AddJournalEntry(NewJournalEntryCodeHash(as.addressHandler, as, as.account.CodeHash))
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

// RootHash returns the hash of the account's (data) root. Nil for library SC
func (as *AccountState) RootHash() []byte {
	return as.account.RootHash
}

// SetRootHash sets the (data) root on the current account. The account will be considered as a data related SC account
func (as *AccountState) SetRootHash(accountsHandler AccountsHandler, root []byte) error {
	if accountsHandler == nil {
		return ErrNilAccountsHandler
	}

	accountsHandler.AddJournalEntry(NewJournalEntryRootHash(as.addressHandler, as, as.account.RootHash))
	as.account.RootHash = root
	return nil
}

// SetRootHashNoJournal sets the root data hash without making a Journal entry
// Should be used only in reverts
func (as *AccountState) SetRootHashNoJournal(rootHash []byte) {
	as.account.RootHash = rootHash
}

// DataTrie returns the trie that holds the current account's data
func (as *AccountState) DataTrie() trie.PatriciaMerkelTree {
	return as.dataTrie
}

// SetDataTrie sets the trie that holds the current account's data
func (as *AccountState) SetDataTrie(trie trie.PatriciaMerkelTree) {
	as.dataTrie = trie
}

// OriginalValue returns the value for a key stored in originalData map which is acting like a cache
func (as *AccountState) OriginalValue(key []byte) []byte {
	return as.originalData[string(key)]
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

// ClearDataCaches empties the dirtyData map and original map
func (as *AccountState) ClearDataCaches() {
	as.dirtyData = make(map[string][]byte)
	as.originalData = make(map[string][]byte)
}

// DirtyData returns the map of (key, value) pairs that contain the data needed to be saved in the data trie
func (as *AccountState) DirtyData() map[string][]byte {
	return as.dirtyData
}

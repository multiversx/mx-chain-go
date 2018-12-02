package state

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
)

// HashLength defines how many bytes are used in a hash
const HashLength = 32

// AccountsHandler is used for the structure that manages the accounts
type AccountsHandler interface {
	AddJournalEntry(je JournalEntry)
	Commit() ([]byte, error)
	GetOrCreateAccountState(addressHandler AddressHandler) (AccountStateHandler, error)
	HasAccountState(addressHandler AddressHandler) (bool, error)
	JournalLen() int
	PutCode(acState AccountStateHandler, code []byte) error
	RemoveAccount(addressHandler AddressHandler) error
	RemoveCode(codeHash []byte) error
	RetrieveDataTrie(acountStateHandler AccountStateHandler) error
	RevertToSnapshot(snapshot int) error
	SaveAccountState(acountStateHandler AccountStateHandler) error
	SaveData(acountStateHandler AccountStateHandler) error
}

// AccountStateHandler models what an AccountState struct should do
type AccountStateHandler interface {
	AddressHandler() AddressHandler
	Balance() big.Int
	Code() []byte
	CodeHash() []byte
	ClearDataCaches()
	DataTrie() trie.PatriciaMerkelTree
	DirtyData() map[string][]byte
	Nonce() uint64
	OriginalValue(key []byte) []byte
	RetrieveValue(key []byte) ([]byte, error)
	RootHash() []byte
	SaveKeyValue(key []byte, value []byte)
	SetBalance(accountsHandler AccountsHandler, value *big.Int) error
	SetBalanceNoJournal(value *big.Int)
	SetCode(code []byte)
	SetCodeHash(accountsHandler AccountsHandler, codeHash []byte) error
	SetCodeHashNoJournal(codeHash []byte)
	SetDataTrie(trie trie.PatriciaMerkelTree)
	SetNonce(accountsHandler AccountsHandler, nonce uint64) error
	SetNonceNoJournal(nonce uint64)
	SetRootHash(accountsHandler AccountsHandler, rootHash []byte) error
	SetRootHashNoJournal(rootHash []byte)
}

// AddressHandler models what an Address struct should do
type AddressHandler interface {
	Bytes() []byte
	Hash(hasher hashing.Hasher) []byte
	//Hex() string
}

// JournalEntry will be used to implement different state changes to be able to easily revert them
type JournalEntry interface {
	Revert(accountsHandler AccountsHandler) error
	DirtiedAddress() AddressHandler
}

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
	PutCode(acState AccountStateHandler, code []byte) error
	RemoveCode(codeHash []byte) error
	RetrieveDataTrie(acState AccountStateHandler) error
	SaveData(acState AccountStateHandler) error
	HasAccountState(address AddressHandler) (bool, error)
	SaveAccountState(acState AccountStateHandler) error
	RemoveAccount(address AddressHandler) error
	GetOrCreateAccountState(address AddressHandler) (AccountStateHandler, error)
	AddJurnalEntry(je JournalEntry)
	RevertFromSnapshot(snapshot int) error
	JournalLen() int
	Commit() ([]byte, error)
}

// AccountStateHandler models what an AccountState struct should do
type AccountStateHandler interface {
	Address() AddressHandler
	Nonce() uint64
	SetNonce(accounts AccountsHandler, nonce uint64) error
	SetNonceNoJournal(nonce uint64)
	Balance() big.Int
	SetBalance(accounts AccountsHandler, value *big.Int) error
	SetBalanceNoJournal(value *big.Int)
	CodeHash() []byte
	SetCodeHash(accounts AccountsHandler, codeHash []byte) error
	SetCodeHashNoJournal(codeHash []byte)
	Code() []byte
	SetCode(code []byte)
	Root() []byte
	SetRoot(accounts AccountsHandler, root []byte) error
	SetRootNoJournal(root []byte)
	DataTrie() trie.PatriciaMerkelTree
	SetDataTrie(t trie.PatriciaMerkelTree)
	RetrieveValue(key []byte) ([]byte, error)
	SaveKeyValue(key []byte, value []byte)
	CollapseDirty()
	ClearDirty()
	DirtyData() map[string][]byte
}

// AddressHandler models what an Address struct should do
type AddressHandler interface {
	Bytes() []byte
	Hash(hasher hashing.Hasher) []byte
	Hex() string
}

// JournalEntry will be used to implement different state changes to be able to easily revert them
type JournalEntry interface {
	Revert(accounts AccountsHandler) error
	DirtyAddress() AddressHandler
}

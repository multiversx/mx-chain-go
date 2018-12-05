package state

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

// HashLength defines how many bytes are used in a hash
const HashLength = 32

// AddressContainer models what an Address struct should do
type AddressContainer interface {
	Bytes() []byte
	Hash() []byte
}

// IntBalancer is an interface for int based balance
type IntBalancer interface {
	Balance() big.Int
	SetBalance(big.Int)
}

// SliceBalancer is an interface for slice based balance (useful when serializing/deserializing) data
type SliceBalancer interface {
	Balance() []byte
	SetBalance([]byte)
}

// BaseAccountContainer is an interface that knows abaot Nonce, Code Hash and Root Hash
type BaseAccountContainer interface {
	Nonce() uint64
	SetNonce(uint64)
	CodeHash() []byte
	SetCodeHash([]byte)
	RootHash() []byte
	SetRootHash([]byte)
}

// DbAccountContainer defines the interface for the account that will actually be saved on a DB
// This structure is Cap'n'Proto compliant
type DbAccountContainer interface {
	BaseAccountContainer

	SliceBalancer
}

// AccountContainer is used for the structure that manipulates Account data
type AccountContainer interface {
	BaseAccountContainer

	IntBalancer
	LoadFromDbAccount(source DbAccountContainer) error
	SaveToDbAccount(target DbAccountContainer) error
}

// SimpleAccountWrapper models what an AccountWrap struct should do
// It knows about code and data, as data structures not hashes
type SimpleAccountWrapper interface {
	AccountContainer

	AddressContainer() AddressContainer
	Code() []byte
	SetCode(code []byte)
	DataTrie() trie.PatriciaMerkelTree
	SetDataTrie(trie trie.PatriciaMerkelTree)
}

// ModifyingDataAccountWrapper models what an AccountWrap struct should do
// It knows how to manipulate data hold by a SC account
type ModifyingDataAccountWrapper interface {
	SimpleAccountWrapper

	ClearDataCaches()
	DirtyData() map[string][]byte
	OriginalValue(key []byte) []byte
	RetrieveValue(key []byte) ([]byte, error)
	SaveKeyValue(key []byte, value []byte)
}

// JournalizedAccountWrapper models what an AccountWrap struct should do
// It knows how to journalize changes
type JournalizedAccountWrapper interface {
	ModifyingDataAccountWrapper

	SetNonceWithJournal(uint64) error
	SetBalanceWithJournal(big.Int) error
	SetCodeHashWithJournal([]byte) error
	SetRootHashWithJournal([]byte) error
}

// AccountsAdapter is used for the structure that manages the accounts on top of a trie.PatriciaMerkleTrie
// implementation
type AccountsAdapter interface {
	AddJournalEntry(je JournalEntry)
	Commit() ([]byte, error)
	GetJournalizedAccount(addressContainer AddressContainer) (JournalizedAccountWrapper, error)
	HasAccount(addressContainer AddressContainer) (bool, error)
	JournalLen() int
	PutCode(journalizedAccountWrapper JournalizedAccountWrapper, code []byte) error
	RemoveAccount(addressContainer AddressContainer) error
	RemoveCode(codeHash []byte) error
	LoadDataTrie(journalizedAccountWrapper JournalizedAccountWrapper) error
	RevertToSnapshot(snapshot int) error
	SaveJournalizedAccount(journalizedAccountWrapper JournalizedAccountWrapper) error
	SaveData(journalizedAccountWrapper JournalizedAccountWrapper) error
	RootHash() []byte
	RecreateTrie(rootHash []byte) error
}

// JournalEntry will be used to implement different state changes to be able to easily revert them
type JournalEntry interface {
	Revert(accountsAdapter AccountsAdapter) error
	DirtiedAddress() AddressContainer
}

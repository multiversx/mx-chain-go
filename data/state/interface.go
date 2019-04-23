package state

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

// HashLength defines how many bytes are used in a hash
const HashLength = 32

// AddressConverter is used to convert to/from AddressContainer
type AddressConverter interface {
	CreateAddressFromPublicKeyBytes(pubKey []byte) (AddressContainer, error)
	ConvertToHex(addressContainer AddressContainer) (string, error)
	CreateAddressFromHex(hexAddress string) (AddressContainer, error)
	PrepareAddressBytes(addressBytes []byte) ([]byte, error)
}

// AddressContainer models what an Address struct should do
type AddressContainer interface {
	Bytes() []byte
}

// AccountFactory creates an account of different types
type AccountFactory interface {
	CreateAccount(address AddressContainer, tracker AccountTracker) AccountWrapper
}

type AccountTracker interface {
	SaveAccount(accountWrapper AccountWrapper) error
	Journalize(entry JournalEntry)
}

type Updater interface {
	Update(key, value []byte) error
}

// AccountWrapper models what an AccountWrap struct should do
// It knows about code and data, as data structures not hashes
type AccountWrapper interface {
	AddressContainer() AddressContainer

	GetCodeHash() []byte
	SetCodeHash([]byte)
	SetCodeHashWithJournal([]byte) error
	GetCode() []byte
	SetCode(code []byte)

	GetRootHash() []byte
	SetRootHash([]byte)
	SetRootHashWithJournal([]byte) error
	DataTrie() trie.PatriciaMerkelTree
	SetDataTrie(trie trie.PatriciaMerkelTree)
	DataTrieTracker() DataTrieTracker
}

// DataTrieTracker models what how to manipulate data held by a SC account
type DataTrieTracker interface {
	ClearDataCaches()
	DirtyData() map[string][]byte
	OriginalValue(key []byte) []byte
	RetrieveValue(key []byte) ([]byte, error)
	SaveKeyValue(key []byte, value []byte)
	SetDataTrie(tr trie.PatriciaMerkelTree)
}

// AccountsAdapter is used for the structure that manages the accounts on top of a trie.PatriciaMerkleTrie
// implementation
type AccountsAdapter interface {
	GetAccountWithJournal(addressContainer AddressContainer) (AccountWrapper, error) // will create if it not exist
	GetExistingAccount(addressContainer AddressContainer) (AccountWrapper, error)
	HasAccount(addressContainer AddressContainer) (bool, error)
	RemoveAccount(addressContainer AddressContainer) error
	Commit() ([]byte, error)
	JournalLen() int
	RevertToSnapshot(snapshot int) error
	RootHash() []byte
	RecreateTrie(rootHash []byte) error
	PutCode(accountWrapper AccountWrapper, code []byte) error
	RemoveCode(codeHash []byte) error
	LoadDataTrie(accountWrapper AccountWrapper) error
	SaveDataTrie(accountWrapper AccountWrapper) error
}

// JournalEntry will be used to implement different state changes to be able to easily revert them
type JournalEntry interface {
	Revert() (AccountWrapper, error)
}

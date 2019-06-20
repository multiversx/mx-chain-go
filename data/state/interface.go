package state

import (
	"github.com/ElrondNetwork/elrond-go/data/trie"
)

// HashLength defines how many bytes are used in a hash
const HashLength = 32

// AddressConverter is used to convert to/from AddressContainer
type AddressConverter interface {
	AddressLen() int
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
	CreateAccount(address AddressContainer, tracker AccountTracker) (AccountHandler, error)
}

// AccountTracker saves an account state and journalizes new entries
type AccountTracker interface {
	SaveAccount(accountHandler AccountHandler) error
	Journalize(entry JournalEntry)
}

// Updater set a new value for a key, implemented by trie
type Updater interface {
	Update(key, value []byte) error
}

// AccountHandler models a state account, which can journalize and revert
// It knows about code and data, as data structures not hashes
type AccountHandler interface {
	AddressContainer() AddressContainer

	GetCodeHash() []byte
	SetCodeHash([]byte)
	SetCodeHashWithJournal([]byte) error
	GetCode() []byte
	SetCode(code []byte)
	SetNonce(nonce uint64)
	GetNonce() uint64
	SetNonceWithJournal(nonce uint64) error

	GetRootHash() []byte
	SetRootHash([]byte)
	SetRootHashWithJournal([]byte) error
	DataTrie() trie.PatriciaMerkelTree
	SetDataTrie(trie trie.PatriciaMerkelTree)
	DataTrieTracker() DataTrieTracker

	IsInterfaceNil() bool
}

// DataTrieTracker models what how to manipulate data held by a SC account
type DataTrieTracker interface {
	ClearDataCaches()
	DirtyData() map[string][]byte
	OriginalValue(key []byte) []byte
	RetrieveValue(key []byte) ([]byte, error)
	SaveKeyValue(key []byte, value []byte)
	SetDataTrie(tr trie.PatriciaMerkelTree)
	DataTrie() trie.PatriciaMerkelTree
}

// AccountsAdapter is used for the structure that manages the accounts on top of a trie.PatriciaMerkleTrie
// implementation
type AccountsAdapter interface {
	GetAccountWithJournal(addressContainer AddressContainer) (AccountHandler, error) // will create if it not exist
	GetExistingAccount(addressContainer AddressContainer) (AccountHandler, error)
	HasAccount(addressContainer AddressContainer) (bool, error)
	RemoveAccount(addressContainer AddressContainer) error
	Commit() ([]byte, error)
	JournalLen() int
	RevertToSnapshot(snapshot int) error
	RootHash() []byte
	RecreateTrie(rootHash []byte) error
	PutCode(accountHandler AccountHandler, code []byte) error
	RemoveCode(codeHash []byte) error
	LoadDataTrie(accountHandler AccountHandler) error
	SaveDataTrie(accountHandler AccountHandler) error
}

// JournalEntry will be used to implement different state changes to be able to easily revert them
type JournalEntry interface {
	Revert() (AccountHandler, error)
}

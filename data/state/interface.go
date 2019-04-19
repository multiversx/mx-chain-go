package state

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

// HashLength defines how many bytes are used in a hash
const HashLength = 32

// RegistrationAddress holds the defined registration address
var RegistrationAddress = NewAddress(make([]byte, 32))

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

// AccountInterface models what an account struct should do
type AccountInterface interface {
	IntegrityAndValidity() error
}

// AccountFactory creates an account of different types
type AccountFactory interface {
	CreateAccount()
}

// AccountWrapper models what an AccountWrap struct should do
// It knows about code and data, as data structures not hashes
type AccountWrapper interface {
	TrackableDataTrieInterface

	AddressContainer() AddressContainer
	Code() []byte
	SetCode(code []byte)
	DataTrie() trie.PatriciaMerkelTree
	SetDataTrie(trie trie.PatriciaMerkelTree)
	AppendRegistrationData(data *RegistrationData) error
	CleanRegistrationData() error
	TrimLastRegistrationData() error

	GetRootHash() []byte
	SetRootHash([]byte)
	GetCodeHash() []byte
	SetCodeHash([]byte)
}

// TrackableDataAccountWrapper models what an AccountWrap struct should do
// It knows how to manipulate data held by a SC account
type TrackableDataTrieInterface interface {
	ClearDataCaches()
	DirtyData() map[string][]byte
	OriginalValue(key []byte) []byte
	RetrieveValue(key []byte) ([]byte, error)
	SaveKeyValue(key []byte, value []byte)
}

// AccountsAdapter is used for the structure that manages the accounts on top of a trie.PatriciaMerkleTrie
// implementation
type AccountsAdapter interface {
	SaveAccount(accountWrapper AccountWrapper) error
	GetAccountWithJournal(addressContainer AddressContainer) (AccountWrapper, error) // will create if it not exist
	GetExistingAccount(addressContainer AddressContainer) (AccountWrapper, error)
	HasAccount(addressContainer AddressContainer) (bool, error)
	RemoveAccount(addressContainer AddressContainer) error

	Commit() ([]byte, error)
	JournalLen() int

	PutCode(accountWrapper AccountWrapper, code []byte) error
	RemoveCode(codeHash []byte) error

	LoadDataTrie(accountWrapper AccountWrapper) error
	RecreateTrie(rootHash []byte) error

	RevertToSnapshot(snapshot int) error

	SaveData(accountWrapper AccountWrapper) error
	RootHash() []byte
}

// JournalEntry will be used to implement different state changes to be able to easily revert them
type JournalEntry interface {
	Revert(accountsAdapter AccountsAdapter) error
	DirtiedAddress() AddressContainer
}

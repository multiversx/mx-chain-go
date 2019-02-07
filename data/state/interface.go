package state

import (
	"math/big"

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

// AccountWrapper models what an AccountWrap struct should do
// It knows about code and data, as data structures not hashes
type AccountWrapper interface {
	BaseAccount() *Account
	AddressContainer() AddressContainer
	Code() []byte
	SetCode(code []byte)
	DataTrie() trie.PatriciaMerkelTree
	SetDataTrie(trie trie.PatriciaMerkelTree)
	AppendRegistrationData(data *RegistrationData) error
	CleanRegistrationData() error
	TrimLastRegistrationData() error
}

// TrackableDataAccountWrapper models what an AccountWrap struct should do
// It knows how to manipulate data held by a SC account
type TrackableDataAccountWrapper interface {
	AccountWrapper

	ClearDataCaches()
	DirtyData() map[string][]byte
	OriginalValue(key []byte) []byte
	RetrieveValue(key []byte) ([]byte, error)
	SaveKeyValue(key []byte, value []byte)
}

// JournalizedAccountWrapper models what an AccountWrap struct should do
// It knows how to journalize changes
type JournalizedAccountWrapper interface {
	TrackableDataAccountWrapper

	SetNonceWithJournal(uint64) error
	SetBalanceWithJournal(*big.Int) error
	SetCodeHashWithJournal([]byte) error
	SetRootHashWithJournal([]byte) error
	AppendDataRegistrationWithJournal(*RegistrationData) error
}

// AccountsAdapter is used for the structure that manages the accounts on top of a trie.PatriciaMerkleTrie
// implementation
type AccountsAdapter interface {
	AddJournalEntry(je JournalEntry)
	Commit() ([]byte, error)
	GetJournalizedAccount(addressContainer AddressContainer) (JournalizedAccountWrapper, error)
	GetExistingAccount(addressContainer AddressContainer) (AccountWrapper, error)
	HasAccount(addressContainer AddressContainer) (bool, error)
	JournalLen() int
	PutCode(journalizedAccountWrapper JournalizedAccountWrapper, code []byte) error
	RemoveAccount(addressContainer AddressContainer) error
	RemoveCode(codeHash []byte) error
	LoadDataTrie(accountWrapper AccountWrapper) error
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

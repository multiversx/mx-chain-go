package state

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// AccountsDbIdentifier is the type of accounts db
type AccountsDbIdentifier byte

const (
	// UserAccountsState is the user accounts
	UserAccountsState AccountsDbIdentifier = 0
	// PeerAccountsState is the peer accounts
	PeerAccountsState AccountsDbIdentifier = 1
)

// HashLength defines how many bytes are used in a hash
const HashLength = 32

// PubkeyConverter can convert public key bytes to/from a human readable form
type PubkeyConverter interface {
	Len() int
	Decode(humanReadable string) ([]byte, error)
	Encode(pkBytes []byte) string
	CreateAddressFromString(humanReadable string) (AddressContainer, error)
	CreateAddressFromBytes(pkBytes []byte) (AddressContainer, error)
	IsInterfaceNil() bool
}

// AddressContainer models what an Address struct should do
type AddressContainer interface {
	Bytes() []byte
	IsInterfaceNil() bool
}

// AccountFactory creates an account of different types
type AccountFactory interface {
	CreateAccount(address AddressContainer) (AccountHandler, error)
	IsInterfaceNil() bool
}

// Updater set a new value for a key, implemented by trie
type Updater interface {
	Update(key, value []byte) error
	IsInterfaceNil() bool
}

// AccountHandler models a state account, which can journalize and revert
// It knows about code and data, as data structures not hashes
type AccountHandler interface {
	AddressContainer() AddressContainer
	IncreaseNonce(nonce uint64)
	GetNonce() uint64
	IsInterfaceNil() bool
}

// PeerAccountHandler models a peer state account, which can journalize a normal account's data
//  with some extra features like signing statistics or rating information
type PeerAccountHandler interface {
	GetBLSPublicKey() []byte
	SetBLSPublicKey([]byte) error
	GetRewardAddress() []byte
	SetRewardAddress([]byte) error
	GetStake() *big.Int
	SetStake(*big.Int) error
	GetAccumulatedFees() *big.Int
	AddToAccumulatedFees(*big.Int)
	GetJailTime() TimePeriod
	SetJailTime(TimePeriod)
	GetList() string
	GetIndex() uint32
	GetCurrentShardId() uint32
	SetCurrentShardId(uint32)
	GetNextShardId() uint32
	SetNextShardId(uint32)
	GetNodeInWaitingList() bool
	SetNodeInWaitingList(bool)
	GetUnStakedNonce() uint64
	SetUnStakedNonce(uint64)
	IncreaseLeaderSuccessRate(uint32)
	DecreaseLeaderSuccessRate(uint32)
	IncreaseValidatorSuccessRate(uint32)
	DecreaseValidatorSuccessRate(uint32)
	GetNumSelectedInSuccessBlocks() uint32
	IncreaseNumSelectedInSuccessBlocks()
	GetLeaderSuccessRate() SignRate
	GetValidatorSuccessRate() SignRate
	GetTotalLeaderSuccessRate() SignRate
	GetTotalValidatorSuccessRate() SignRate
	SetListAndIndex(shardID uint32, list string, index uint32)
	GetRating() uint32
	SetRating(uint32)
	GetTempRating() uint32
	SetTempRating(uint32)
	GetConsecutiveProposerMisses() uint32
	SetConsecutiveProposerMisses(uint322 uint32)
	ResetAtNewEpoch()
	AccountHandler
}

// UserAccountHandler models a user account, which can journalize account's data with some extra features
// like balance, developer rewards, owner
type UserAccountHandler interface {
	SetCode(code []byte)
	GetCode() []byte
	SetCodeHash([]byte)
	GetCodeHash() []byte
	SetRootHash([]byte)
	GetRootHash() []byte
	SetDataTrie(trie data.Trie)
	DataTrie() data.Trie
	DataTrieTracker() DataTrieTracker
	AddToBalance(value *big.Int) error
	SubFromBalance(value *big.Int) error
	GetBalance() *big.Int
	ClaimDeveloperRewards([]byte) (*big.Int, error)
	AddToDeveloperReward(*big.Int)
	GetDeveloperReward() *big.Int
	ChangeOwnerAddress([]byte, []byte) error
	SetOwnerAddress([]byte)
	GetOwnerAddress() []byte
	AccountHandler
}

// DataTrieTracker models what how to manipulate data held by a SC account
type DataTrieTracker interface {
	ClearDataCaches()
	DirtyData() map[string][]byte
	OriginalValue(key []byte) []byte
	RetrieveValue(key []byte) ([]byte, error)
	SaveKeyValue(key []byte, value []byte)
	SetDataTrie(tr data.Trie)
	DataTrie() data.Trie
	IsInterfaceNil() bool
}

// AccountsAdapter is used for the structure that manages the accounts on top of a trie.PatriciaMerkleTrie
// implementation
type AccountsAdapter interface {
	GetExistingAccount(addressContainer AddressContainer) (AccountHandler, error)
	LoadAccount(address AddressContainer) (AccountHandler, error)
	SaveAccount(account AccountHandler) error
	RemoveAccount(addressContainer AddressContainer) error
	Commit() ([]byte, error)
	JournalLen() int
	RevertToSnapshot(snapshot int) error

	RootHash() ([]byte, error)
	RecreateTrie(rootHash []byte) error
	PruneTrie(rootHash []byte, identifier data.TriePruningIdentifier)
	CancelPrune(rootHash []byte, identifier data.TriePruningIdentifier)
	SnapshotState(rootHash []byte)
	SetStateCheckpoint(rootHash []byte)
	IsPruningEnabled() bool
	GetAllLeaves(rootHash []byte) (map[string][]byte, error)
	RecreateAllTries(rootHash []byte) (map[string]data.Trie, error)
	IsInterfaceNil() bool
}

// JournalEntry will be used to implement different state changes to be able to easily revert them
type JournalEntry interface {
	Revert() (AccountHandler, error)
	IsInterfaceNil() bool
}

// TriesHolder is used to store multiple tries
type TriesHolder interface {
	Put([]byte, data.Trie)
	Replace(key []byte, tr data.Trie)
	Get([]byte) data.Trie
	GetAll() []data.Trie
	Reset()
	IsInterfaceNil() bool
}

type baseAccountHandler interface {
	AddressContainer() AddressContainer
	IncreaseNonce(nonce uint64)
	GetNonce() uint64
	SetCode(code []byte)
	GetCode() []byte
	SetCodeHash([]byte)
	GetCodeHash() []byte
	SetRootHash([]byte)
	GetRootHash() []byte
	SetDataTrie(trie data.Trie)
	DataTrie() data.Trie
	DataTrieTracker() DataTrieTracker
	IsInterfaceNil() bool
}

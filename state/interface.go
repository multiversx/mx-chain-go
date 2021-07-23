package state

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
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

// AccountFactory creates an account of different types
type AccountFactory interface {
	CreateAccount(address []byte) (vmcommon.AccountHandler, error)
	IsInterfaceNil() bool
}

// Updater set a new value for a key, implemented by trie
type Updater interface {
	Get(key []byte) ([]byte, error)
	Update(key, value []byte) error
	IsInterfaceNil() bool
}

// PeerAccountHandler models a peer state account, which can journalize a normal account's data
//  with some extra features like signing statistics or rating information
type PeerAccountHandler interface {
	GetBLSPublicKey() []byte
	SetBLSPublicKey([]byte) error
	GetRewardAddress() []byte
	SetRewardAddress([]byte) error
	GetAccumulatedFees() *big.Int
	AddToAccumulatedFees(*big.Int)
	GetList() string
	GetIndexInList() uint32
	GetShardId() uint32
	SetUnStakedEpoch(epoch uint32)
	GetUnStakedEpoch() uint32
	IncreaseLeaderSuccessRate(uint32)
	DecreaseLeaderSuccessRate(uint32)
	IncreaseValidatorSuccessRate(uint32)
	DecreaseValidatorSuccessRate(uint32)
	IncreaseValidatorIgnoredSignaturesRate(uint32)
	GetNumSelectedInSuccessBlocks() uint32
	IncreaseNumSelectedInSuccessBlocks()
	GetLeaderSuccessRate() SignRate
	GetValidatorSuccessRate() SignRate
	GetValidatorIgnoredSignaturesRate() uint32
	GetTotalLeaderSuccessRate() SignRate
	GetTotalValidatorSuccessRate() SignRate
	GetTotalValidatorIgnoredSignaturesRate() uint32
	SetListAndIndex(shardID uint32, list string, index uint32)
	GetRating() uint32
	SetRating(uint32)
	GetTempRating() uint32
	SetTempRating(uint32)
	GetConsecutiveProposerMisses() uint32
	SetConsecutiveProposerMisses(uint322 uint32)
	ResetAtNewEpoch()
	vmcommon.AccountHandler
}

// UserAccountHandler models a user account, which can journalize account's data with some extra features
// like balance, developer rewards, owner
type UserAccountHandler interface {
	SetCode(code []byte)
	SetCodeMetadata(codeMetadata []byte)
	GetCodeMetadata() []byte
	SetCodeHash([]byte)
	GetCodeHash() []byte
	SetRootHash([]byte)
	GetRootHash() []byte
	SetDataTrie(trie temporary.Trie)
	DataTrie() temporary.Trie
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
	SetUserName(userName []byte)
	GetUserName() []byte
	vmcommon.AccountHandler
}

// DataTrieTracker models what how to manipulate data held by a SC account
type DataTrieTracker interface {
	ClearDataCaches()
	DirtyData() map[string][]byte
	RetrieveValue(key []byte) ([]byte, error)
	SaveKeyValue(key []byte, value []byte) error
	SetDataTrie(tr temporary.Trie)
	DataTrie() temporary.Trie
	IsInterfaceNil() bool
}

// AccountsAdapter is used for the structure that manages the accounts on top of a trie.PatriciaMerkleTrie
// implementation
type AccountsAdapter interface {
	GetExistingAccount(address []byte) (vmcommon.AccountHandler, error)
	LoadAccount(address []byte) (vmcommon.AccountHandler, error)
	SaveAccount(account vmcommon.AccountHandler) error
	RemoveAccount(address []byte) error
	Commit() ([]byte, error)
	JournalLen() int
	RevertToSnapshot(snapshot int) error
	GetNumCheckpoints() uint32
	GetCode(codeHash []byte) []byte
	RootHash() ([]byte, error)
	RecreateTrie(rootHash []byte) error
	PruneTrie(rootHash []byte, identifier temporary.TriePruningIdentifier)
	CancelPrune(rootHash []byte, identifier temporary.TriePruningIdentifier)
	SnapshotState(rootHash []byte)
	SetStateCheckpoint(rootHash []byte)
	IsPruningEnabled() bool
	GetAllLeaves(rootHash []byte) (chan core.KeyValueHolder, error)
	RecreateAllTries(rootHash []byte) (map[string]temporary.Trie, error)
	GetTrie(rootHash []byte) (temporary.Trie, error)
	Close() error
	IsInterfaceNil() bool
}

// JournalEntry will be used to implement different state changes to be able to easily revert them
type JournalEntry interface {
	Revert() (vmcommon.AccountHandler, error)
	IsInterfaceNil() bool
}

// TriesHolder is used to store multiple tries
type TriesHolder interface {
	Put([]byte, temporary.Trie)
	Replace(key []byte, tr temporary.Trie)
	Get([]byte) temporary.Trie
	GetAll() []temporary.Trie
	Reset()
	IsInterfaceNil() bool
}

type baseAccountHandler interface {
	AddressBytes() []byte
	IncreaseNonce(nonce uint64)
	GetNonce() uint64
	SetCode(code []byte)
	HasNewCode() bool
	SetCodeMetadata(codeMetadata []byte)
	GetCodeMetadata() []byte
	SetCodeHash([]byte)
	GetCodeHash() []byte
	SetRootHash([]byte)
	GetRootHash() []byte
	SetDataTrie(trie temporary.Trie)
	DataTrie() temporary.Trie
	DataTrieTracker() DataTrieTracker
	IsInterfaceNil() bool
}

// AccountsDBImporter is used in importing accounts
type AccountsDBImporter interface {
	ImportAccount(account vmcommon.AccountHandler) error
	Commit() ([]byte, error)
	IsInterfaceNil() bool
}

// DBRemoveCacher is used to cache keys that will be deleted from the database
type DBRemoveCacher interface {
	Put([]byte, temporary.ModifiedHashes) error
	Evict([]byte) (temporary.ModifiedHashes, error)
	ShouldKeepHash(hash string, identifier temporary.TriePruningIdentifier) (bool, error)
	IsInterfaceNil() bool
	Close() error
}

// AtomicBuffer is used to buffer byteArrays
type AtomicBuffer interface {
	Add(rootHash []byte)
	RemoveAll() [][]byte
	Len() int
}

// StoragePruningManager is used to manage all state pruning operations
type StoragePruningManager interface {
	MarkForEviction([]byte, []byte, temporary.ModifiedHashes, temporary.ModifiedHashes) error
	PruneTrie(rootHash []byte, identifier temporary.TriePruningIdentifier, tsm temporary.StorageManager)
	CancelPrune(rootHash []byte, identifier temporary.TriePruningIdentifier, tsm temporary.StorageManager)
	Close() error
	IsInterfaceNil() bool
}

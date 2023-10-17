package state

import (
	"context"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// AccountFactory creates an account of different types
type AccountFactory interface {
	CreateAccount(address []byte) (vmcommon.AccountHandler, error)
	IsInterfaceNil() bool
}

// Updater set a new value for a key, implemented by trie
type Updater interface {
	Get(key []byte) ([]byte, error)
	Put(key, val []byte) error
	Remove(key []byte) error
	IsInterfaceNil() bool
}

// Updater set a new value for a key, implemented by trie
type TrieUpdater interface {
	Update(key, value []byte) error
	IsInterfaceNil() bool
}

// AccountsAdapter is used for the structure that manages the accounts on top of a trie.PatriciaMerkleTrie
// implementation
type AccountsAdapter interface {
	GetExistingAccount(address []byte) (vmcommon.AccountHandler, error)
	GetAccountFromBytes(address []byte, accountBytes []byte) (vmcommon.AccountHandler, error)
	LoadAccount(address []byte) (vmcommon.AccountHandler, error)
	SaveAccount(account vmcommon.AccountHandler) error
	RemoveAccount(address []byte) error
	CommitInEpoch(currentEpoch uint32, epochToCommit uint32) ([]byte, error)
	Commit() ([]byte, error)
	JournalLen() int
	RevertToSnapshot(snapshot int) error
	GetCode(codeHash []byte) []byte
	RootHash() ([]byte, error)
	RecreateTrie(rootHash []byte) error
	RecreateTrieFromEpoch(options common.RootHashHolder) error
	PruneTrie(rootHash []byte, identifier TriePruningIdentifier, handler PruningHandler)
	CancelPrune(rootHash []byte, identifier TriePruningIdentifier)
	SnapshotState(rootHash []byte, epoch uint32)
	SetStateCheckpoint(rootHash []byte)
	IsPruningEnabled() bool
	GetAllLeaves(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, trieLeafParser common.TrieLeafParser) error
	RecreateAllTries(rootHash []byte) (map[string]common.Trie, error)
	GetTrie(rootHash []byte) (common.Trie, error)
	GetStackDebugFirstEntry() []byte
	SetSyncer(syncer AccountsDBSyncer) error
	StartSnapshotIfNeeded() error
	Close() error
	IsInterfaceNil() bool
}

// SnapshotsManager defines the methods for the snapshot manager
type SnapshotsManager interface {
	SnapshotState(rootHash []byte, epoch uint32, trieStorageManager common.StorageManager)
	SetStateCheckpoint(rootHash []byte, trieStorageManager common.StorageManager)
	StartSnapshotAfterRestartIfNeeded(trieStorageManager common.StorageManager) error
	IsSnapshotInProgress() bool
	SetSyncer(syncer AccountsDBSyncer) error
	IsInterfaceNil() bool
}

// StateMetrics defines the methods for the state metrics
type StateMetrics interface {
	UpdateMetricsOnSnapshotStart()
	UpdateMetricsOnSnapshotCompletion(stats common.SnapshotStatisticsHandler)
	GetSnapshotMessage() string
	IsInterfaceNil() bool
}

// IteratorChannelsProvider defines the methods for the iterator channels provider
type IteratorChannelsProvider interface {
	GetIteratorChannels() *common.TrieIteratorChannels
	IsInterfaceNil() bool
}

// AccountsAdapterWithClean extends the AccountsAdapter interface with a CleanCache method
type AccountsAdapterWithClean interface {
	AccountsAdapter
	CleanCache()
}

// AccountsDBSyncer defines the methods for the accounts db syncer
type AccountsDBSyncer interface {
	SyncAccounts(rootHash []byte, storageMarker common.StorageMarker) error
	IsInterfaceNil() bool
}

// AccountsRepository handles the defined execution based on the query options
type AccountsRepository interface {
	GetAccountWithBlockInfo(address []byte, options api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error)
	GetCodeWithBlockInfo(codeHash []byte, options api.AccountQueryOptions) ([]byte, common.BlockInfo, error)
	GetCurrentStateAccountsWrapper() AccountsAdapterAPI
	Close() error
	IsInterfaceNil() bool
}

// JournalEntry will be used to implement different state changes to be able to easily revert them
type JournalEntry interface {
	Revert() (vmcommon.AccountHandler, error)
	IsInterfaceNil() bool
}

type baseAccountHandler interface {
	AddressBytes() []byte
	IncreaseNonce(nonce uint64)
	GetNonce() uint64
	SetCode(code []byte)
	GetCode() []byte
	HasNewCode() bool
	SetCodeMetadata(codeMetadata []byte)
	GetCodeMetadata() []byte
	SetCodeHash([]byte)
	GetCodeHash() []byte
	SetRootHash([]byte)
	GetRootHash() []byte
	SetDataTrie(trie common.Trie)
	DataTrie() common.DataTrieHandler
	SaveDirtyData(trie common.Trie) ([]core.TrieData, error)
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
	Put([]byte, common.ModifiedHashes) error
	Evict([]byte) (common.ModifiedHashes, error)
	ShouldKeepHash(hash string, identifier TriePruningIdentifier) (bool, error)
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
	MarkForEviction([]byte, []byte, common.ModifiedHashes, common.ModifiedHashes) error
	PruneTrie(rootHash []byte, identifier TriePruningIdentifier, tsm common.StorageManager, handler PruningHandler)
	CancelPrune(rootHash []byte, identifier TriePruningIdentifier, tsm common.StorageManager)
	Close() error
	IsInterfaceNil() bool
}

// PruningHandler defines different options for pruning
type PruningHandler interface {
	IsPruningEnabled() bool
}

// BlockInfoProvider defines the behavior of a struct able to provide the block information used in state tries
type BlockInfoProvider interface {
	GetBlockInfo() common.BlockInfo
	IsInterfaceNil() bool
}

// AccountsAdapterAPI defines the extension of the AccountsAdapter that should be used in API calls
type AccountsAdapterAPI interface {
	AccountsAdapter
	GetAccountWithBlockInfo(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error)
	GetCodeWithBlockInfo(codeHash []byte, options common.RootHashHolder) ([]byte, common.BlockInfo, error)
}

// DataTrie defines the behavior of a data trie
type DataTrie interface {
	common.Trie

	UpdateWithVersion(key []byte, value []byte, version core.TrieNodeVersion) error
	CollectLeavesForMigration(args vmcommon.ArgsMigrateDataTrieLeaves) error
}

// PeerAccountHandler models a peer state account, which can journalize a normal account's data
//  with some extra features like signing statistics or rating information
type PeerAccountHandler interface {
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
	SetDataTrie(trie common.Trie)
	DataTrie() common.DataTrieHandler
	RetrieveValue(key []byte) ([]byte, uint32, error)
	SaveKeyValue(key []byte, value []byte) error
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
	IsGuarded() bool
	GetAllLeaves(leavesChannels *common.TrieIteratorChannels, ctx context.Context) error
	vmcommon.AccountHandler
}

// DataTrieTracker models how to manipulate data held by a SC account
type DataTrieTracker interface {
	RetrieveValue(key []byte) ([]byte, uint32, error)
	SaveKeyValue(key []byte, value []byte) error
	SetDataTrie(tr common.Trie)
	DataTrie() common.DataTrieHandler
	SaveDirtyData(common.Trie) ([]core.TrieData, error)
	MigrateDataTrieLeaves(args vmcommon.ArgsMigrateDataTrieLeaves) error
	IsInterfaceNil() bool
}

// SignRate defines the operations of an entity that holds signing statistics
type SignRate interface {
	GetNumSuccess() uint32
	GetNumFailure() uint32
}

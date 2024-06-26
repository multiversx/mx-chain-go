package state

import (
	"context"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/api"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/common"
)

// AccountFactory creates an account of different types
type AccountFactory interface {
	CreateAccount(address []byte) (vmcommon.AccountHandler, error)
	IsInterfaceNil() bool
}

// Updater set a new value for a key, implemented by trie
type Updater interface {
	Get(key []byte) ([]byte, uint32, error)
	Update(key, value []byte) error
	IsInterfaceNil() bool
}

// PeerAccountHandler models a peer state account, which can journalize a normal account's data
// with some extra features like signing statistics or rating information
type PeerAccountHandler interface {
	GetBLSPublicKey() []byte
	SetBLSPublicKey([]byte) error
	GetRewardAddress() []byte
	SetRewardAddress([]byte) error
	GetAccumulatedFees() *big.Int
	AddToAccumulatedFees(*big.Int)
	GetList() string
	GetPreviousList() string
	GetIndexInList() uint32
	GetPreviousIndexInList() uint32
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
	SetListAndIndex(shardID uint32, list string, index uint32, updatePreviousValues bool)
	GetRating() uint32
	SetRating(uint32)
	GetTempRating() uint32
	SetTempRating(uint32)
	GetConsecutiveProposerMisses() uint32
	SetConsecutiveProposerMisses(uint322 uint32)
	ResetAtNewEpoch()
	SetPreviousList(list string)
	vmcommon.AccountHandler
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

// StateStatsHandler defines the behaviour needed to handler state statistics
type StateStatsHandler interface {
	ResetSnapshot()
	SnapshotStats() []string
	IsInterfaceNil() bool
}

// LastSnapshotMarker manages the lastSnapshot marker operations
type LastSnapshotMarker interface {
	AddMarker(trieStorageManager common.StorageManager, epoch uint32, rootHash []byte)
	RemoveMarker(trieStorageManager common.StorageManager, epoch uint32, rootHash []byte)
	GetMarkerInfo(trieStorageManager common.StorageManager) ([]byte, error)
	IsInterfaceNil() bool
}

// ShardValidatorsInfoMapHandler shall be used to manage operations inside
// a <shardID, []ValidatorInfoHandler> map in a concurrent-safe way.
type ShardValidatorsInfoMapHandler interface {
	GetShardValidatorsInfoMap() map[uint32][]ValidatorInfoHandler
	GetAllValidatorsInfo() []ValidatorInfoHandler
	GetValidator(blsKey []byte) ValidatorInfoHandler

	Add(validator ValidatorInfoHandler) error
	Delete(validator ValidatorInfoHandler) error
	DeleteByKey(blsKey []byte, shardID uint32)
	Replace(old ValidatorInfoHandler, new ValidatorInfoHandler) error
	ReplaceValidatorByKey(oldBlsKey []byte, new ValidatorInfoHandler, shardID uint32) bool
	SetValidatorsInShard(shardID uint32, validators []ValidatorInfoHandler) error
	SetValidatorsInShardUnsafe(shardID uint32, validators []ValidatorInfoHandler)
}

// ValidatorInfoHandler defines which data shall a validator info hold.
type ValidatorInfoHandler interface {
	IsInterfaceNil() bool

	GetPublicKey() []byte
	GetShardId() uint32
	GetList() string
	GetIndex() uint32
	GetPreviousIndex() uint32
	GetTempRating() uint32
	GetRating() uint32
	GetRatingModifier() float32
	GetRewardAddress() []byte
	GetLeaderSuccess() uint32
	GetLeaderFailure() uint32
	GetValidatorSuccess() uint32
	GetValidatorFailure() uint32
	GetValidatorIgnoredSignatures() uint32
	GetNumSelectedInSuccessBlocks() uint32
	GetAccumulatedFees() *big.Int
	GetTotalLeaderSuccess() uint32
	GetTotalLeaderFailure() uint32
	GetTotalValidatorSuccess() uint32
	GetTotalValidatorFailure() uint32
	GetTotalValidatorIgnoredSignatures() uint32
	GetPreviousList() string

	SetPublicKey(publicKey []byte)
	SetShardId(shardID uint32)
	SetPreviousList(list string)
	SetList(list string)
	SetIndex(index uint32)
	SetListAndIndex(list string, index uint32, updatePreviousValues bool)
	SetTempRating(tempRating uint32)
	SetRating(rating uint32)
	SetRatingModifier(ratingModifier float32)
	SetRewardAddress(rewardAddress []byte)
	SetLeaderSuccess(leaderSuccess uint32)
	SetLeaderFailure(leaderFailure uint32)
	SetValidatorSuccess(validatorSuccess uint32)
	SetValidatorFailure(validatorFailure uint32)
	SetValidatorIgnoredSignatures(validatorIgnoredSignatures uint32)
	SetNumSelectedInSuccessBlocks(numSelectedInSuccessBlock uint32)
	SetAccumulatedFees(accumulatedFees *big.Int)
	SetTotalLeaderSuccess(totalLeaderSuccess uint32)
	SetTotalLeaderFailure(totalLeaderFailure uint32)
	SetTotalValidatorSuccess(totalValidatorSuccess uint32)
	SetTotalValidatorFailure(totalValidatorFailure uint32)
	SetTotalValidatorIgnoredSignatures(totalValidatorIgnoredSignatures uint32)

	ShallowClone() ValidatorInfoHandler
	String() string
	GoString() string
}

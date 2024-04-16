package common

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	crypto "github.com/multiversx/mx-chain-crypto-go"
)

// TrieIteratorChannels defines the channels that are being used when iterating the trie nodes
type TrieIteratorChannels struct {
	LeavesChan chan core.KeyValueHolder
	ErrChan    BufferedErrChan
}

// TrieType defines the type of the trie
type TrieType string

const (
	// MainTrie represents the main trie in which all the accounts and SC code are stored
	MainTrie TrieType = "mainTrie"

	// DataTrie represents a data trie in which all the data related to an account is stored
	DataTrie TrieType = "dataTrie"
)

// BufferedErrChan is an interface that defines the methods for a buffered error channel
type BufferedErrChan interface {
	WriteInChanNonBlocking(err error)
	ReadFromChanNonBlocking() error
	Close()
	IsInterfaceNil() bool
}

// Trie is an interface for Merkle Trees implementations
type Trie interface {
	Get(key []byte) ([]byte, uint32, error)
	Update(key, value []byte) error
	Delete(key []byte) error
	RootHash() ([]byte, error)
	Commit() error
	Recreate(options RootHashHolder) (Trie, error)
	String() string
	GetObsoleteHashes() [][]byte
	GetDirtyHashes() (ModifiedHashes, error)
	GetOldRoot() []byte
	GetSerializedNodes([]byte, uint64) ([][]byte, uint64, error)
	GetSerializedNode([]byte) ([]byte, error)
	GetAllLeavesOnChannel(allLeavesChan *TrieIteratorChannels, ctx context.Context, rootHash []byte, keyBuilder KeyBuilder, trieLeafParser TrieLeafParser) error
	GetAllHashes() ([][]byte, error)
	GetProof(key []byte) ([][]byte, []byte, error)
	VerifyProof(rootHash []byte, key []byte, proof [][]byte) (bool, error)
	GetStorageManager() StorageManager
	IsMigratedToLatestVersion() (bool, error)
	Close() error
	IsInterfaceNil() bool
}

// TrieLeafParser is used to parse trie leaves
type TrieLeafParser interface {
	ParseLeaf(key []byte, val []byte, version core.TrieNodeVersion) (core.KeyValueHolder, error)
	IsInterfaceNil() bool
}

// TrieStats is used to collect the trie statistics for the given rootHash
type TrieStats interface {
	GetTrieStats(address string, rootHash []byte) (TrieStatisticsHandler, error)
}

// StorageMarker is used to mark the given storer as synced and active
type StorageMarker interface {
	MarkStorerAsSyncedAndActive(storer StorageManager)
	IsInterfaceNil() bool
}

// KeyBuilder is used for building trie keys as you traverse the trie
type KeyBuilder interface {
	BuildKey(keyPart []byte)
	GetKey() ([]byte, error)
	DeepClone() KeyBuilder
	ShallowClone() KeyBuilder
	Size() uint
	IsInterfaceNil() bool
}

// DataTrieHandler is an interface that declares the methods used for dataTries
type DataTrieHandler interface {
	RootHash() ([]byte, error)
	GetAllLeavesOnChannel(leavesChannels *TrieIteratorChannels, ctx context.Context, rootHash []byte, keyBuilder KeyBuilder, trieLeafParser TrieLeafParser) error
	IsMigratedToLatestVersion() (bool, error)
	IsInterfaceNil() bool
}

// StorageManager manages all trie storage operations
type StorageManager interface {
	TrieStorageInteractor
	GetFromCurrentEpoch(key []byte) ([]byte, error)
	PutInEpoch(key []byte, val []byte, epoch uint32) error
	PutInEpochWithoutCache(key []byte, val []byte, epoch uint32) error
	TakeSnapshot(address string, rootHash []byte, mainTrieRootHash []byte, iteratorChannels *TrieIteratorChannels, missingNodesChan chan []byte, stats SnapshotStatisticsHandler, epoch uint32)
	GetLatestStorageEpoch() (uint32, error)
	IsPruningEnabled() bool
	IsPruningBlocked() bool
	EnterPruningBufferingMode()
	ExitPruningBufferingMode()
	RemoveFromAllActiveEpochs(hash []byte) error
	SetEpochForPutOperation(uint32)
	ShouldTakeSnapshot() bool
	IsSnapshotSupported() bool
	GetBaseTrieStorageManager() StorageManager
	IsClosed() bool
	Close() error
	IsInterfaceNil() bool
}

// TrieStorageInteractor defines the methods used for interacting with the trie storage
type TrieStorageInteractor interface {
	BaseStorer
	GetIdentifier() string
	GetStateStatsHandler() StateStatisticsHandler
}

// BaseStorer define the base methods needed for a storer
type BaseStorer interface {
	Put(key, val []byte) error
	Get(key []byte) ([]byte, error)
	Remove(key []byte) error
	Close() error
	IsInterfaceNil() bool
}

// SnapshotDbHandler is used to keep track of how many references a snapshot db has
type SnapshotDbHandler interface {
	BaseStorer
	IsInUse() bool
	DecreaseNumReferences()
	IncreaseNumReferences()
	MarkForRemoval()
	MarkForDisconnection()
	SetPath(string)
}

// TriesHolder is used to store multiple tries
type TriesHolder interface {
	Put([]byte, Trie)
	Replace(key []byte, tr Trie)
	Get([]byte) Trie
	GetAll() []Trie
	Reset()
	IsInterfaceNil() bool
}

// Locker defines the operations used to lock different critical areas. Implemented by the RWMutex.
type Locker interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

// MerkleProofVerifier is used to verify merkle proofs
type MerkleProofVerifier interface {
	VerifyProof(rootHash []byte, key []byte, proof [][]byte) (bool, error)
}

// SizeSyncStatisticsHandler extends the SyncStatisticsHandler interface by allowing setting up the trie node size
type SizeSyncStatisticsHandler interface {
	data.SyncStatisticsHandler
	AddNumBytesReceived(bytes uint64)
	NumBytesReceived() uint64
	NumTries() int
	AddProcessingTime(duration time.Duration)
	IncrementIteration()
	ProcessingTime() time.Duration
	NumIterations() int
}

// SnapshotStatisticsHandler is used to measure different statistics for the trie snapshot
type SnapshotStatisticsHandler interface {
	SnapshotFinished()
	NewSnapshotStarted()
	WaitForSnapshotsToFinish()
	AddTrieStats(handler TrieStatisticsHandler, trieType TrieType)
	GetSnapshotDuration() int64
	GetSnapshotNumNodes() uint64
	IsInterfaceNil() bool
}

// TrieStatisticsHandler is used to collect different statistics about a single trie
type TrieStatisticsHandler interface {
	AddBranchNode(level int, size uint64)
	AddExtensionNode(level int, size uint64)
	AddLeafNode(level int, size uint64, version core.TrieNodeVersion)
	AddAccountInfo(address string, rootHash []byte)

	GetTotalNodesSize() uint64
	GetTotalNumNodes() uint64
	GetMaxTrieDepth() uint32
	GetBranchNodesSize() uint64
	GetNumBranchNodes() uint64
	GetExtensionNodesSize() uint64
	GetNumExtensionNodes() uint64
	GetLeafNodesSize() uint64
	GetNumLeafNodes() uint64
	GetLeavesMigrationStats() map[core.TrieNodeVersion]uint64

	MergeTriesStatistics(statsToBeMerged TrieStatisticsHandler)
	ToString() []string
	IsInterfaceNil() bool
}

// TriesStatisticsCollector is used to merge the statistics for multiple tries
type TriesStatisticsCollector interface {
	Add(trieStats TrieStatisticsHandler, trieType TrieType)
	Print()
	GetNumNodes() uint64
}

// StateStatisticsHandler defines the behaviour of a storage statistics handler
type StateStatisticsHandler interface {
	Reset()
	ResetSnapshot()

	IncrementCache()
	Cache() uint64
	IncrementSnapshotCache()
	SnapshotCache() uint64

	IncrementPersister(epoch uint32)
	Persister(epoch uint32) uint64
	IncrementSnapshotPersister(epoch uint32)
	SnapshotPersister(epoch uint32) uint64

	IncrementTrie()
	Trie() uint64

	ProcessingStats() []string
	SnapshotStats() []string

	IsInterfaceNil() bool
}

// ProcessStatusHandler defines the behavior of a component able to hold the current status of the node and
// able to tell if the node is idle or processing/committing a block
type ProcessStatusHandler interface {
	SetBusy(reason string)
	SetIdle()
	IsIdle() bool
	IsInterfaceNil() bool
}

// BlockInfo provides a block information such as nonce, hash, roothash and so on
type BlockInfo interface {
	GetNonce() uint64
	GetHash() []byte
	GetRootHash() []byte
	Equal(blockInfo BlockInfo) bool
	IsInterfaceNil() bool
}

// ReceiptsHolder holds receipts content (e.g. miniblocks)
type ReceiptsHolder interface {
	GetMiniblocks() []*block.MiniBlock
	IsInterfaceNil() bool
}

// RootHashHolder holds a rootHash and the corresponding epoch
type RootHashHolder interface {
	GetRootHash() []byte
	GetEpoch() core.OptionalUint32
	String() string
	IsInterfaceNil() bool
}

// GasScheduleNotifierAPI defines the behavior of the gas schedule notifier components that is used for api
type GasScheduleNotifierAPI interface {
	core.GasScheduleNotifier
	LatestGasScheduleCopy() map[string]map[string]uint64
}

// PidQueueHandler defines the behavior of a queue of pids
type PidQueueHandler interface {
	Push(pid core.PeerID)
	Pop() core.PeerID
	IndexOf(pid core.PeerID) int
	Promote(idx int)
	Remove(pid core.PeerID)
	DataSizeInBytes() int
	Get(idx int) core.PeerID
	Len() int
	IsInterfaceNil() bool
}

// EnableEpochsHandler is used to verify which flags are set in a specific epoch based on EnableEpochs config
type EnableEpochsHandler interface {
	GetCurrentEpoch() uint32
	IsFlagDefined(flag core.EnableEpochFlag) bool
	IsFlagEnabled(flag core.EnableEpochFlag) bool
	IsFlagEnabledInEpoch(flag core.EnableEpochFlag, epoch uint32) bool
	GetActivationEpoch(flag core.EnableEpochFlag) uint32

	IsInterfaceNil() bool
}

// ManagedPeersHolder defines the operations of an entity that holds managed identities for a node
type ManagedPeersHolder interface {
	AddManagedPeer(privateKeyBytes []byte) error
	GetPrivateKey(pkBytes []byte) (crypto.PrivateKey, error)
	GetP2PIdentity(pkBytes []byte) ([]byte, core.PeerID, error)
	GetMachineID(pkBytes []byte) (string, error)
	GetNameAndIdentity(pkBytes []byte) (string, string, error)
	IncrementRoundsWithoutReceivedMessages(pkBytes []byte)
	ResetRoundsWithoutReceivedMessages(pkBytes []byte, pid core.PeerID)
	GetManagedKeysByCurrentNode() map[string]crypto.PrivateKey
	GetLoadedKeysByCurrentNode() [][]byte
	IsKeyManagedByCurrentNode(pkBytes []byte) bool
	IsKeyRegistered(pkBytes []byte) bool
	IsPidManagedByCurrentNode(pid core.PeerID) bool
	IsKeyValidator(pkBytes []byte) bool
	SetValidatorState(pkBytes []byte, state bool)
	GetNextPeerAuthenticationTime(pkBytes []byte) (time.Time, error)
	SetNextPeerAuthenticationTime(pkBytes []byte, nextTime time.Time)
	IsMultiKeyMode() bool
	GetRedundancyStepInReason() string
	IsInterfaceNil() bool
}

// MissingTrieNodesNotifier defines the operations of an entity that notifies about missing trie nodes
type MissingTrieNodesNotifier interface {
	RegisterHandler(handler StateSyncNotifierSubscriber) error
	AsyncNotifyMissingTrieNode(hash []byte)
	IsInterfaceNil() bool
}

// StateSyncNotifierSubscriber defines the operations of an entity that subscribes to a missing trie nodes notifier
type StateSyncNotifierSubscriber interface {
	MissingDataTrieNodeFound(hash []byte)
	IsInterfaceNil() bool
}

// ManagedPeersMonitor defines the operations of an entity that monitors the managed peers holder
type ManagedPeersMonitor interface {
	GetManagedKeysCount() int
	GetManagedKeys() [][]byte
	GetLoadedKeys() [][]byte
	GetEligibleManagedKeys() ([][]byte, error)
	GetWaitingManagedKeys() ([][]byte, error)
	IsInterfaceNil() bool
}

// TxExecutionOrderHandler is used to collect and provide the order of transactions execution
type TxExecutionOrderHandler interface {
	Add(txHash []byte)
	GetItemAtIndex(index uint32) ([]byte, error)
	GetOrder(txHash []byte) (int, error)
	Remove(txHash []byte)
	RemoveMultiple(txHashes [][]byte)
	GetItems() [][]byte
	Contains(txHash []byte) bool
	Clear()
	Len() int
	IsInterfaceNil() bool
}

// ExecutionOrderGetter defines the functionality of a component that can return the execution order of a block transactions
type ExecutionOrderGetter interface {
	GetItemAtIndex(index uint32) ([]byte, error)
	GetOrder(txHash []byte) (int, error)
	GetItems() [][]byte
	Contains(txHash []byte) bool
	Len() int
	IsInterfaceNil() bool
}

// TrieNodeData is used to retrieve the data of a trie node
type TrieNodeData interface {
	GetKeyBuilder() KeyBuilder
	GetData() []byte
	Size() uint64
	IsLeaf() bool
}

// DfsIterator is used to iterate the trie nodes in a depth-first search manner
type DfsIterator interface {
	GetLeaves(numLeaves int, ctx context.Context) (map[string]string, error)
	GetIteratorId() []byte
	Clone() DfsIterator
	FinishedIteration() bool
	Size() uint64
	IsInterfaceNil() bool
}

// TrieLeavesRetriever is used to retrieve the leaves from the trie. If there is a saved checkpoint for the iterator id,
// it will continue to iterate from the checkpoint.
type TrieLeavesRetriever interface {
	GetLeaves(numLeaves int, rootHash []byte, iteratorID []byte, ctx context.Context) (map[string]string, []byte, error)
	IsInterfaceNil() bool
}

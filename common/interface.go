package common

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
)

// NumNodesDTO represents the DTO structure that will hold the number of nodes split by category and other
// trie structure relevant data such as maximum number of trie levels including the roothash node and all leaves
type NumNodesDTO struct {
	Leaves     int
	Extensions int
	Branches   int
	MaxLevel   int
}

// Trie is an interface for Merkle Trees implementations
type Trie interface {
	Get(key []byte) ([]byte, error)
	Update(key, value []byte) error
	Delete(key []byte) error
	RootHash() ([]byte, error)
	Commit() error
	Recreate(root []byte) (Trie, error)
	RecreateFromEpoch(options RootHashHolder) (Trie, error)
	String() string
	GetObsoleteHashes() [][]byte
	GetDirtyHashes() (ModifiedHashes, error)
	GetOldRoot() []byte
	GetSerializedNodes([]byte, uint64) ([][]byte, uint64, error)
	GetSerializedNode([]byte) ([]byte, error)
	GetNumNodes() NumNodesDTO
	GetAllLeavesOnChannel(leavesChannel chan core.KeyValueHolder, ctx context.Context, rootHash []byte) error
	GetAllHashes() ([][]byte, error)
	GetProof(key []byte) ([][]byte, []byte, error)
	VerifyProof(rootHash []byte, key []byte, proof [][]byte) (bool, error)
	GetStorageManager() StorageManager
	MarkStorerAsSyncedAndActive()
	Close() error
	IsInterfaceNil() bool
}

// StorageManager manages all trie storage operations
type StorageManager interface {
	Get(key []byte) ([]byte, error)
	GetFromEpoch(key []byte, epoch uint32) ([]byte, error)
	GetFromCurrentEpoch(key []byte) ([]byte, error)
	Put(key []byte, val []byte) error
	PutInEpoch(key []byte, val []byte, epoch uint32) error
	PutInEpochWithoutCache(key []byte, val []byte, epoch uint32) error
	TakeSnapshot(rootHash []byte, mainTrieRootHash []byte, leavesChan chan core.KeyValueHolder, errChan chan error, stats SnapshotStatisticsHandler, epoch uint32)
	SetCheckpoint(rootHash []byte, mainTrieRootHash []byte, leavesChan chan core.KeyValueHolder, errChan chan error, stats SnapshotStatisticsHandler)
	GetLatestStorageEpoch() (uint32, error)
	IsPruningEnabled() bool
	IsPruningBlocked() bool
	EnterPruningBufferingMode()
	ExitPruningBufferingMode()
	AddDirtyCheckpointHashes([]byte, ModifiedHashes) bool
	Remove(hash []byte) error
	SetEpochForPutOperation(uint32)
	ShouldTakeSnapshot() bool
	IsClosed() bool
	Close() error
	IsInterfaceNil() bool
}

// DBWriteCacher is used to cache changes made to the trie, and only write to the database when it's needed
type DBWriteCacher interface {
	Put(key, val []byte) error
	Get(key []byte) ([]byte, error)
	Remove(key []byte) error
	Close() error
	IsInterfaceNil() bool
}

// SnapshotDbHandler is used to keep track of how many references a snapshot db has
type SnapshotDbHandler interface {
	DBWriteCacher
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
	AddSize(uint64)
	SnapshotFinished()
	NewSnapshotStarted()
	NewDataTrie()
	WaitForSnapshotsToFinish()
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

// RootHashHolder holds a rootHash
type RootHashHolder interface {
	GetRootHash() []byte
	GetEpoch() core.OptionalUint32
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

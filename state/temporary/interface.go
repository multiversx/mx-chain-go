package temporary

import (
	"context"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/config"
)

//TODO: move this interfaces in the packages where they are used

// TriePruningIdentifier is the type for trie pruning identifiers
type TriePruningIdentifier byte

const (
	// OldRoot is appended to the key when oldHashes are added to the evictionWaitingList
	OldRoot TriePruningIdentifier = 0
	// NewRoot is appended to the key when newHashes are added to the evictionWaitingList
	NewRoot TriePruningIdentifier = 1
)

// ModifiedHashes is used to memorize all old hashes and new hashes from when a trie is committed
type ModifiedHashes map[string]struct{}

// Clone is used to create a clone of the map
func (mh ModifiedHashes) Clone() ModifiedHashes {
	newMap := make(ModifiedHashes)
	for key := range mh {
		newMap[key] = struct{}{}
	}

	return newMap
}

// NumNodesDTO represents the DTO structure that will hold the number of nodes split by category and other
// trie structure relevant data such as maximum number of trie levels including the roothash node and all leaves
type NumNodesDTO struct {
	Leaves     int
	Extensions int
	Branches   int
	MaxLevel   int
}

// TrieSyncer synchronizes the trie, asking on the network for the missing nodes
type TrieSyncer interface {
	StartSyncing(rootHash []byte, ctx context.Context) error
	IsInterfaceNil() bool
}

// TrieCreateArgs holds arguments for calling the Create method on the TrieFactory
type TrieCreateArgs struct {
	TrieStorageConfig  config.StorageConfig
	ShardID            string
	PruningEnabled     bool
	CheckpointsEnabled bool
	MaxTrieLevelInMem  uint
}

//Trie is an interface for Merkle Trees implementations
type Trie interface {
	Get(key []byte) ([]byte, error)
	Update(key, value []byte) error
	Delete(key []byte) error
	RootHash() ([]byte, error)
	Commit() error
	Recreate(root []byte) (Trie, error)
	String() string
	GetObsoleteHashes() [][]byte
	GetDirtyHashes() (ModifiedHashes, error)
	GetOldRoot() []byte
	GetSerializedNodes([]byte, uint64) ([][]byte, uint64, error)
	GetSerializedNode([]byte) ([]byte, error)
	GetNumNodes() NumNodesDTO
	GetAllLeavesOnChannel(rootHash []byte) (chan core.KeyValueHolder, error)
	GetAllHashes() ([][]byte, error)
	GetProof(key []byte) ([][]byte, error)
	VerifyProof(key []byte, proof [][]byte) (bool, error)
	GetStorageManager() StorageManager
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

// StorageManager manages all trie storage operations
type StorageManager interface {
	Database() DBWriteCacher
	TakeSnapshot([]byte, bool, chan core.KeyValueHolder)
	SetCheckpoint([]byte, chan core.KeyValueHolder)
	GetSnapshotThatContainsHash(rootHash []byte) SnapshotDbHandler
	IsPruningEnabled() bool
	IsPruningBlocked() bool
	EnterPruningBufferingMode()
	ExitPruningBufferingMode()
	GetSnapshotDbBatchDelay() int
	AddDirtyCheckpointHashes([]byte, ModifiedHashes) bool
	Remove(hash []byte) error
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

// SyncStatisticsHandler defines the methods for a component able to store the sync statistics for a trie
type SyncStatisticsHandler interface {
	Reset()
	AddNumReceived(value int)
	AddNumLarge(value int)
	SetNumMissing(rootHash []byte, value int)
	NumReceived() int
	NumLarge() int
	NumMissing() int
	IsInterfaceNil() bool
}

// CheckpointHashesHolder is used to hold the hashes that need to be committed in the future state checkpoint
type CheckpointHashesHolder interface {
	Put(rootHash []byte, hashes ModifiedHashes) bool
	RemoveCommitted(lastCommittedRootHash []byte)
	Remove(hash []byte)
	ShouldCommit(hash []byte) bool
	IsInterfaceNil() bool
}

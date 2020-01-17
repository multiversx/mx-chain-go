package update

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// StateSyncer interface defines the methods needed to sync and get all states
type StateSyncer interface {
	GetMetaBlock() (*block.MetaBlock, error)
	SyncAllState(epoch uint32) error
	GetAllTries() (map[string]data.Trie, error)
	GetAllTransactions() (map[string]data.TransactionHandler, error)
	GetAllMiniBlocks() (map[string]*block.MiniBlock, error)
	IsInterfaceNil() bool
}

// TrieSyncer synchronizes the trie, asking on the network for the missing nodes
type TrieSyncer interface {
	StartSyncing(rootHash []byte) error
	Trie() data.Trie
	IsInterfaceNil() bool
}

// TrieSyncContainer keep a list of TrieSyncer
type TrieSyncContainer interface {
	Get(key string) (TrieSyncer, error)
	Add(key string, val TrieSyncer) error
	AddMultiple(keys []string, interceptors []TrieSyncer) error
	Replace(key string, val TrieSyncer) error
	Remove(key string)
	Len() int
	IsInterfaceNil() bool
}

// EpochStartVerifier defines the functionality needed by sync all state from epochTrigger
type EpochStartVerifier interface {
	IsEpochStart() bool
	ReceivedHeader(header data.HeaderHandler)
	Epoch() uint32
	EpochStartMetaHdrHash() []byte
	IsInterfaceNil() bool
}

// HistoryStorer provides storage services in a two layered storage construct, where the first layer is
// represented by a cache and second layer by a persitent storage (DB-like)
type HistoryStorer interface {
	Put(key, data []byte) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) error
	Remove(key []byte) error
	ClearCache()
	DestroyUnit() error
	GetFromEpoch(key []byte, epoch uint32) ([]byte, error)
	HasInEpoch(key []byte, epoch uint32) error

	IsInterfaceNil() bool
}

// MultiFileWriter defines the methods to write in multiple files
type MultiFileWriter interface {
	NewFile(name string) error
	Write(fileName string, key string, value []byte) error
	IsInterfaceNil() bool
}

// MultiFileReader defines the methods to read from multiple files
type MultiFileReader interface {
	GetFileNames() []string
	ReadNextItem(fileName string) (string, []byte, error)
	IsInterfaceNil() bool
}

// HeaderSyncHandler defines the methods to sync and get the epoch start metablock
type HeaderSyncHandler interface {
	SyncEpochStartMetaHeader(epoch uint32, waitTime time.Duration) (*block.MetaBlock, error)
	GetMetaBlock() (*block.MetaBlock, error)
	IsInterfaceNil() bool
}

// EpochStartTriesSyncHandler defines the methods to sync all tries from a given epoch start metablock
type EpochStartTriesSyncHandler interface {
	SyncTriesFrom(meta *block.MetaBlock, waitTime time.Duration) error
	GetTries() (map[string]data.Trie, error)
	IsInterfaceNil() bool
}

// EpochStartPendingMiniBlocksSyncHandler defines the methods to sync all pending miniblocks
type EpochStartPendingMiniBlocksSyncHandler interface {
	SyncPendingMiniBlocksFromMeta(meta *block.MetaBlock, waitTime time.Duration) error
	GetMiniBlocks() (map[string]*block.MiniBlock, error)
	IsInterfaceNil() bool
}

// PendingTransactionsSyncHandler defines the methods to sync all transactions from a set of miniblocks
type PendingTransactionsSyncHandler interface {
	SyncPendingTransactionsFor(miniBlocks map[string]*block.MiniBlock, epoch uint32, waitTime time.Duration) error
	GetTransactions() (map[string]data.TransactionHandler, error)
	IsInterfaceNil() bool
}

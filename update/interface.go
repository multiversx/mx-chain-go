package update

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
)

type StateSyncer interface {
	GetMetaBlock() *block.MetaBlock
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

// TrieSyncerContainer keep a list of TrieSyncer
type TrieSyncContainer interface {
	Get(key string) (TrieSyncer, error)
	Add(key string, val TrieSyncer) error
	AddMultiple(keys []string, interceptors []TrieSyncer) error
	Replace(key string, val TrieSyncer) error
	Remove(key string)
	Len() int
	IsInterfaceNil() bool
}

// DataTriesContainer is used to store multiple tries
type DataTriesContainer interface {
	Put([]byte, data.Trie)
	Get([]byte) data.Trie
	GetAllTries() map[string]data.Trie
	IsInterfaceNil() bool
}

// EpochStartNotifier defines which actions should be done for handling new epoch's events
type EpochStartNotifier interface {
	RegisterHandler(handler notifier.SubscribeFunctionHandler)
	IsInterfaceNil() bool
}

// EpochStartVerifier defines the functionality needed by sync all state from epochTrigger
type EpochStartVerifier interface {
	IsEpochStart() bool
	ReceivedHeader(header data.HeaderHandler)
	Epoch() uint32
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

// MultiFileWriter writes the pushed data in several files in a buffered way
type MultiFileWriter interface {
	NewFile(name string) error
	Write(fileName string, key string, value []byte) error
	Finish()
	IsInterfaceNil() bool
}

// MultiFileReaders reads data from several files in a buffered way
type MultiFileReader interface {
	GetFileNames() []string
	ReadNextItem(fileName string) (string, []byte, error)
	Finish()
	IsInterfaceNil() bool
}

// RequestHandler defines the methods through which request to data can be made
type RequestHandler interface {
	RequestTransaction(shardId uint32, txHashes [][]byte)
	RequestUnsignedTransactions(destShardID uint32, scrHashes [][]byte)
	RequestRewardTransactions(destShardID uint32, txHashes [][]byte)
	RequestMiniBlock(shardId uint32, miniblockHash []byte)
	RequestStartOfEpochMetaBlock(epoch uint32)
	RequestShardHeader(shardId uint32, hash []byte)
	RequestMetaHeader(hash []byte)
	RequestMetaHeaderByNonce(nonce uint64)
	RequestShardHeaderByNonce(shardId uint32, nonce uint64)
	RequestTrieNodes(shardId uint32, hash []byte)
	IsInterfaceNil() bool
}

// ExportHandler defines the methods to export the current state of the blockchain
type ExportHandler interface {
	ExportAll(epoch uint32) error
	IsInterfaceNil() bool
}

// ImportHandler defines the methods to import the full state of the blockchain
type ImportHandler interface {
	ImportAll() error
	GetAllGenesisBlocks() map[uint32]data.HeaderHandler
	IsInterfaceNil() bool
}

package update

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
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

// AccountsHandlerContainer keep a list of AccountsAdapters
type AccountsHandlerContainer interface {
	Get(key string) (state.AccountsAdapter, error)
	Add(key string, val state.AccountsAdapter) error
	AddMultiple(keys []string, interceptors []state.AccountsAdapter) error
	Replace(key string, val state.AccountsAdapter) error
	Remove(key string)
	Len() int
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

type MultiFileWriter interface {
	NewFile(name string) error
	Write(fileName string, key string, value []byte) error
	IsInterfaceNil() bool
}

type MultiFileReader interface {
	GetFileNames() []string
	ReadNextItem(fileName string) (string, []byte, error)
	IsInterfaceNil() bool
}

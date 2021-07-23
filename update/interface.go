package update

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
)

// StateSyncer interface defines the methods needed to sync and get all states
type StateSyncer interface {
	GetEpochStartMetaBlock() (*block.MetaBlock, error)
	GetUnFinishedMetaBlocks() (map[string]*block.MetaBlock, error)
	SyncAllState(epoch uint32) error
	GetAllTries() (map[string]temporary.Trie, error)
	GetAllTransactions() (map[string]data.TransactionHandler, error)
	GetAllMiniBlocks() (map[string]*block.MiniBlock, error)
	IsInterfaceNil() bool
}

// TrieSyncer synchronizes the trie, asking on the network for the missing nodes
type TrieSyncer interface {
	StartSyncing(rootHash []byte, ctx context.Context) error
	Trie() temporary.Trie
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
	RequestTrieNodes(destShardID uint32, hashes [][]byte, topic string)
	RequestInterval() time.Duration
	SetNumPeersToQuery(key string, intra int, cross int) error
	GetNumPeersToQuery(key string) (int, int, error)
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
	GetValidatorAccountsDB() state.AccountsAdapter
	GetMiniBlocks() map[string]*block.MiniBlock
	GetHardForkMetaBlock() *block.MetaBlock
	GetUnFinishedMetaBlocks() map[string]*block.MetaBlock
	GetTransactions() map[string]data.TransactionHandler
	GetAccountsDBForShard(shardID uint32) state.AccountsAdapter
	Close() error
	IsInterfaceNil() bool
}

// HardForkBlockProcessor defines the methods to process after hardfork
type HardForkBlockProcessor interface {
	CreateBlock(body *block.Body, chainID string, round uint64, nonce uint64, epoch uint32) (data.HeaderHandler, error)
	CreateBody() (*block.Body, []*MbInfo, error)
	CreatePostMiniBlocks(mbsInfo []*MbInfo) (*block.Body, []*MbInfo, error)
	IsInterfaceNil() bool
}

// PendingTransactionProcessor defines the methods to process a transaction destination me
type PendingTransactionProcessor interface {
	ProcessTransactionsDstMe(mbInfo *MbInfo) (*block.MiniBlock, error)
	RootHash() ([]byte, error)
	Commit() ([]byte, error)
	IsInterfaceNil() bool
}

// HeaderSyncHandler defines the methods to sync and get the epoch start metablock
type HeaderSyncHandler interface {
	SyncUnFinishedMetaHeaders(epoch uint32) error
	GetEpochStartMetaBlock() (*block.MetaBlock, error)
	GetUnFinishedMetaBlocks() (map[string]*block.MetaBlock, error)
	IsInterfaceNil() bool
}

// EpochStartTriesSyncHandler defines the methods to sync all tries from a given epoch start metablock
type EpochStartTriesSyncHandler interface {
	SyncTriesFrom(meta *block.MetaBlock) error
	GetTries() (map[string]temporary.Trie, error)
	IsInterfaceNil() bool
}

// EpochStartPendingMiniBlocksSyncHandler defines the methods to sync all pending miniblocks
type EpochStartPendingMiniBlocksSyncHandler interface {
	SyncPendingMiniBlocksFromMeta(epochStart *block.MetaBlock, unFinished map[string]*block.MetaBlock, ctx context.Context) error
	GetMiniBlocks() (map[string]*block.MiniBlock, error)
	IsInterfaceNil() bool
}

// PendingTransactionsSyncHandler defines the methods to sync all transactions from a set of miniblocks
type PendingTransactionsSyncHandler interface {
	SyncPendingTransactionsFor(miniBlocks map[string]*block.MiniBlock, epoch uint32, ctx context.Context) error
	GetTransactions() (map[string]data.TransactionHandler, error)
	IsInterfaceNil() bool
}

// MissingHeadersByHashSyncer defines the methods to sync all missing headers by hash
type MissingHeadersByHashSyncer interface {
	SyncMissingHeadersByHash(shardIDs []uint32, headersHashes [][]byte, ctx context.Context) error
	GetHeaders() (map[string]data.HeaderHandler, error)
	ClearFields()
	IsInterfaceNil() bool
}

// DataWriter defines the methods to write data
type DataWriter interface {
	WriteString(s string) (int, error)
	Flush() error
}

// WhiteListHandler is the interface needed to add whitelisted data
type WhiteListHandler interface {
	Remove(keys [][]byte)
	Add(keys [][]byte)
	IsWhiteListed(interceptedData process.InterceptedData) bool
	IsWhiteListedAtLeastOne(identifiers [][]byte) bool
	IsInterfaceNil() bool
}

// AccountsDBSyncer defines the methods for the accounts db syncer
type AccountsDBSyncer interface {
	GetSyncedTries() map[string]temporary.Trie
	SyncAccounts(rootHash []byte) error
	IsInterfaceNil() bool
}

// AccountsDBSyncContainer keep a list of TrieSyncer
type AccountsDBSyncContainer interface {
	Get(key string) (AccountsDBSyncer, error)
	Add(key string, val AccountsDBSyncer) error
	AddMultiple(keys []string, values []AccountsDBSyncer) error
	Replace(key string, val AccountsDBSyncer) error
	Remove(key string)
	Len() int
	IsInterfaceNil() bool
}

// SigVerifier is used to verify the signature on a provided message
type SigVerifier interface {
	Verify(message []byte, sig []byte, pk []byte) error
	IsInterfaceNil() bool
}

// EpochHandler defines the functionality to get the current epoch
type EpochHandler interface {
	MetaEpoch() uint32
	ForceEpochStart(round uint64)
	IsInterfaceNil() bool
}

// ExportFactoryHandler defines the functionality to create and export handler
type ExportFactoryHandler interface {
	Create() (ExportHandler, error)
	IsInterfaceNil() bool
}

// EpochChangeConfirmedNotifier defines the functionality needed to register for the epoch change confirmed event
type EpochChangeConfirmedNotifier interface {
	RegisterForEpochChangeConfirmed(handler func(epoch uint32))
	IsInterfaceNil() bool
}

// Closer defines the functionality of an entity that can be closed
type Closer interface {
	Close() error
	IsInterfaceNil() bool
}

// ImportStartHandler can manage the process of starting the import after the hardfork event
type ImportStartHandler interface {
	IsAfterExportBeforeImport() bool
	ShouldStartImport() bool
	ResetStartImport() error
	SetStartImport() error
	IsInterfaceNil() bool
}

// HardforkStorer manages the export and import of data
type HardforkStorer interface {
	Write(identifier string, key []byte, value []byte) error
	FinishedIdentifier(identifier string) error
	RangeKeys(handler func(identifier string, keys [][]byte) bool)
	Get(identifier string, key []byte) ([]byte, error)
	Close() error
	IsInterfaceNil() bool
}

// GenesisNodesSetupHandler returns the genesis nodes info
type GenesisNodesSetupHandler interface {
	InitialNodesInfoForShard(shardId uint32) ([]sharding.GenesisNodeInfoHandler, []sharding.GenesisNodeInfoHandler, error)
	InitialNodesInfo() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler)
	GetStartTime() int64
	GetRoundDuration() uint64
	GetShardConsensusGroupSize() uint32
	GetMetaConsensusGroupSize() uint32
	MinNumberOfShardNodes() uint32
	MinNumberOfMetaNodes() uint32
	GetHysteresis() float32
	GetAdaptivity() bool
	NumberOfShards() uint32
	MinNumberOfNodes() uint32
	IsInterfaceNil() bool
}

// RoundHandler defines the actions which should be handled by a round implementation
type RoundHandler interface {
	Index() int64
	TimeStamp() time.Time
	IsInterfaceNil() bool
}

// PreferredPeersHolderHandler defines the behavior of a component able to handle preferred peers operations
type PreferredPeersHolderHandler interface {
	Put(publicKey []byte, peerID core.PeerID, shardID uint32)
	Get() map[uint32][]core.PeerID
	Contains(peerID core.PeerID) bool
	Remove(peerID core.PeerID)
	Clear()
	IsInterfaceNil() bool
}

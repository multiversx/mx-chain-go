package dataRetriever

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/counting"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// UnitType is the type for Storage unit identifiers
type UnitType uint8

// String returns the friendly name of the unit
func (ut UnitType) String() string {
	switch ut {
	case TransactionUnit:
		return "TransactionUnit"
	case MiniBlockUnit:
		return "MiniBlockUnit"
	case PeerChangesUnit:
		return "PeerChangesUnit"
	case BlockHeaderUnit:
		return "BlockHeaderUnit"
	case MetaBlockUnit:
		return "MetaBlockUnit"
	case UnsignedTransactionUnit:
		return "UnsignedTransactionUnit"
	case RewardTransactionUnit:
		return "RewardTransactionUnit"
	case MetaHdrNonceHashDataUnit:
		return "MetaHdrNonceHashDataUnit"
	case HeartbeatUnit:
		return "HeartbeatUnit"
	case BootstrapUnit:
		return "BootstrapUnit"
	case StatusMetricsUnit:
		return "StatusMetricsUnit"
	case ReceiptsUnit:
		return "ReceiptsUnit"
	case TrieEpochRootHashUnit:
		return "TrieEpochRootHashUnit"
	}

	if ut < ShardHdrNonceHashDataUnit {
		return fmt.Sprintf("unknown type %d", ut)
	}

	return fmt.Sprintf("%s%d", "ShardHdrNonceHashDataUnit", ut-ShardHdrNonceHashDataUnit)
}

const (
	// TransactionUnit is the transactions storage unit identifier
	TransactionUnit UnitType = 0
	// MiniBlockUnit is the transaction block body storage unit identifier
	MiniBlockUnit UnitType = 1
	// PeerChangesUnit is the peer change block body storage unit identifier
	PeerChangesUnit UnitType = 2
	// BlockHeaderUnit is the Block Headers Storage unit identifier
	BlockHeaderUnit UnitType = 3
	// MetaBlockUnit is the metachain blocks storage unit identifier
	MetaBlockUnit UnitType = 4
	// UnsignedTransactionUnit is the unsigned transaction unit identifier
	UnsignedTransactionUnit UnitType = 5
	// RewardTransactionUnit is the reward transaction unit identifier
	RewardTransactionUnit UnitType = 6
	// MetaHdrNonceHashDataUnit is the meta header nonce-hash pair data unit identifier
	MetaHdrNonceHashDataUnit UnitType = 7
	// HeartbeatUnit is the heartbeat storage unit identifier
	HeartbeatUnit UnitType = 8
	// BootstrapUnit is the bootstrap storage unit identifier
	BootstrapUnit UnitType = 9
	//StatusMetricsUnit is the status metrics storage unit identifier
	StatusMetricsUnit UnitType = 10
	// TxLogsUnit is the transactions logs storage unit identifier
	TxLogsUnit UnitType = 11
	// MiniblocksMetadataUnit is the miniblocks metadata storage unit identifier
	MiniblocksMetadataUnit UnitType = 12
	// EpochByHashUnit is the epoch by hash storage unit identifier
	EpochByHashUnit UnitType = 13
	// MiniblockHashByTxHashUnit is the miniblocks hash by tx hash storage unit identifier
	MiniblockHashByTxHashUnit UnitType = 14
	// ReceiptsUnit is the receipts storage unit identifier
	ReceiptsUnit UnitType = 15
	// ResultsHashesByTxHashUnit is the results hashes by transaction storage unit identifier
	ResultsHashesByTxHashUnit UnitType = 16
	// TrieEpochRootHashUnit is the trie epoch <-> root hash storage unit identifier
	TrieEpochRootHashUnit UnitType = 17

	// ShardHdrNonceHashDataUnit is the header nonce-hash pair data unit identifier
	//TODO: Add only unit types lower than 100
	ShardHdrNonceHashDataUnit UnitType = 100
	//TODO: Do not add unit type greater than 100 as the metachain creates this kind of unit type for each shard.
	//100 -> shard 0, 101 -> shard 1 and so on. This should be replaced with a factory which will manage the unit types
	//creation
)

// ResolverThrottler can monitor the number of the currently running resolver go routines
type ResolverThrottler interface {
	CanProcess() bool
	StartProcessing()
	EndProcessing()
	IsInterfaceNil() bool
}

// Resolver defines what a data resolver should do
type Resolver interface {
	RequestDataFromHash(hash []byte, epoch uint32) error
	ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
	SetResolverDebugHandler(handler ResolverDebugHandler) error
	SetNumPeersToQuery(intra int, cross int)
	NumPeersToQuery() (int, int)
	Close() error
	IsInterfaceNil() bool
}

// TrieNodesResolver defines what a trie nodes resolver should do
type TrieNodesResolver interface {
	Resolver
	RequestDataFromHashArray(hashes [][]byte, epoch uint32) error
}

// HeaderResolver defines what a block header resolver should do
type HeaderResolver interface {
	Resolver
	RequestDataFromNonce(nonce uint64, epoch uint32) error
	RequestDataFromEpoch(identifier []byte) error
	SetEpochHandler(epochHandler EpochHandler) error
}

// MiniBlocksResolver defines what a mini blocks resolver should do
type MiniBlocksResolver interface {
	Resolver
	RequestDataFromHashArray(hashes [][]byte, epoch uint32) error
}

// TopicResolverSender defines what sending operations are allowed for a topic resolver
type TopicResolverSender interface {
	SendOnRequestTopic(rd *RequestData, originalHashes [][]byte) error
	Send(buff []byte, peer core.PeerID) error
	RequestTopic() string
	TargetShardID() uint32
	SetNumPeersToQuery(intra int, cross int)
	SetResolverDebugHandler(handler ResolverDebugHandler) error
	ResolverDebugHandler() ResolverDebugHandler
	NumPeersToQuery() (int, int)
	IsInterfaceNil() bool
}

// ResolversContainer defines a resolvers holder data type with basic functionality
type ResolversContainer interface {
	Get(key string) (Resolver, error)
	Add(key string, val Resolver) error
	AddMultiple(keys []string, resolvers []Resolver) error
	Replace(key string, val Resolver) error
	Remove(key string)
	Len() int
	ResolverKeys() string
	Iterate(handler func(key string, resolver Resolver) bool)
	Close() error
	IsInterfaceNil() bool
}

// ResolversFinder extends a container resolver and have 2 additional functionality
type ResolversFinder interface {
	ResolversContainer
	IntraShardResolver(baseTopic string) (Resolver, error)
	MetaChainResolver(baseTopic string) (Resolver, error)
	CrossShardResolver(baseTopic string, crossShard uint32) (Resolver, error)
	MetaCrossShardResolver(baseTopic string, crossShard uint32) (Resolver, error)
}

// ResolversContainerFactory defines the functionality to create a resolvers container
type ResolversContainerFactory interface {
	Create() (ResolversContainer, error)
	IsInterfaceNil() bool
}

// EpochHandler defines the functionality to get the current epoch
type EpochHandler interface {
	MetaEpoch() uint32
	IsInterfaceNil() bool
}

// ManualEpochStartNotifier can manually notify an epoch change
type ManualEpochStartNotifier interface {
	NewEpoch(epoch uint32)
	CurrentEpoch() uint32
	IsInterfaceNil() bool
}

// EpochProviderByNonce defines the functionality needed for calculating an epoch based on nonce
type EpochProviderByNonce interface {
	EpochForNonce(nonce uint64) (uint32, error)
	IsInterfaceNil() bool
}

// MessageHandler defines the functionality needed by structs to send data to other peers
type MessageHandler interface {
	ConnectedPeersOnTopic(topic string) []core.PeerID
	ConnectedFullHistoryPeersOnTopic(topic string) []core.PeerID
	SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error
	ID() core.PeerID
	IsInterfaceNil() bool
}

// TopicHandler defines the functionality needed by structs to manage topics and message processors
type TopicHandler interface {
	HasTopic(name string) bool
	CreateTopic(name string, createChannelForTopic bool) error
	RegisterMessageProcessor(topic string, identifier string, handler p2p.MessageProcessor) error
}

// TopicMessageHandler defines the functionality needed by structs to manage topics, message processors and to send data
// to other peers
type TopicMessageHandler interface {
	MessageHandler
	TopicHandler
}

// Messenger defines which methods a p2p messenger should implement
type Messenger interface {
	MessageHandler
	TopicHandler
	UnregisterMessageProcessor(topic string, identifier string) error
	UnregisterAllMessageProcessors() error
	UnjoinAllTopics() error
	ConnectedPeers() []core.PeerID
}

// IntRandomizer interface provides functionality over generating integer numbers
type IntRandomizer interface {
	Intn(n int) int
	IsInterfaceNil() bool
}

// StorageType defines the storage levels on a node
type StorageType uint8

// PeerListCreator is used to create a peer list
type PeerListCreator interface {
	PeerList() []core.PeerID
	IntraShardPeerList() []core.PeerID
	FullHistoryList() []core.PeerID
	IsInterfaceNil() bool
}

// ShardedDataCacherNotifier defines what a sharded-data structure can perform
type ShardedDataCacherNotifier interface {
	RegisterOnAdded(func(key []byte, value interface{}))
	ShardDataStore(cacheId string) (c storage.Cacher)
	AddData(key []byte, data interface{}, sizeInBytes int, cacheId string)
	SearchFirstData(key []byte) (value interface{}, ok bool)
	RemoveData(key []byte, cacheId string)
	RemoveSetOfDataFromPool(keys [][]byte, cacheId string)
	ImmunizeSetOfDataAgainstEviction(keys [][]byte, cacheId string)
	RemoveDataFromAllShards(key []byte)
	MergeShardStores(sourceCacheID, destCacheID string)
	Clear()
	ClearShardStore(cacheId string)
	GetCounts() counting.CountsWithSize
	IsInterfaceNil() bool
}

// ShardIdHashMap represents a map for shardId and hash
type ShardIdHashMap interface {
	Load(shardId uint32) ([]byte, bool)
	Store(shardId uint32, hash []byte)
	Range(f func(shardId uint32, hash []byte) bool)
	Delete(shardId uint32)
	IsInterfaceNil() bool
}

// HeadersPool defines what a headers pool structure can perform
type HeadersPool interface {
	Clear()
	AddHeader(headerHash []byte, header data.HeaderHandler)
	RemoveHeaderByHash(headerHash []byte)
	RemoveHeaderByNonceAndShardId(headerNonce uint64, shardId uint32)
	GetHeadersByNonceAndShardId(headerNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error)
	GetHeaderByHash(hash []byte) (data.HeaderHandler, error)
	RegisterHandler(handler func(headerHandler data.HeaderHandler, headerHash []byte))
	Nonces(shardId uint32) []uint64
	Len() int
	MaxSize() int
	IsInterfaceNil() bool
	GetNumHeaders(shardId uint32) int
}

// TransactionCacher defines the methods for the local cacher, info for current round
type TransactionCacher interface {
	Clean()
	GetTx(txHash []byte) (data.TransactionHandler, error)
	AddTx(txHash []byte, tx data.TransactionHandler)
	IsInterfaceNil() bool
}

// PoolsHolder defines getters for data pools
type PoolsHolder interface {
	Transactions() ShardedDataCacherNotifier
	UnsignedTransactions() ShardedDataCacherNotifier
	RewardTransactions() ShardedDataCacherNotifier
	Headers() HeadersPool
	MiniBlocks() storage.Cacher
	PeerChangesBlocks() storage.Cacher
	TrieNodes() storage.Cacher
	SmartContracts() storage.Cacher
	CurrentBlockTxs() TransactionCacher
	IsInterfaceNil() bool
}

// StorageService is the interface for data storage unit provided services
type StorageService interface {
	// GetStorer returns the storer from the chain map
	GetStorer(unitType UnitType) storage.Storer
	// AddStorer will add a new storer to the chain map
	AddStorer(key UnitType, s storage.Storer)
	// Has returns true if the key is found in the selected Unit or false otherwise
	Has(unitType UnitType, key []byte) error
	// Get returns the value for the given key if found in the selected storage unit, nil otherwise
	Get(unitType UnitType, key []byte) ([]byte, error)
	// Put stores the key, value pair in the selected storage unit
	Put(unitType UnitType, key []byte, value []byte) error
	// SetEpochForPutOperation will set the epoch which will be used for the put operation
	SetEpochForPutOperation(epoch uint32)
	// GetAll gets all the elements with keys in the keys array, from the selected storage unit
	// If there is a missing key in the unit, it returns an error
	GetAll(unitType UnitType, keys [][]byte) (map[string][]byte, error)
	// Destroy removes the underlying files/resources used by the storage service
	Destroy() error
	//CloseAll will close all the units
	CloseAll() error
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// DataPacker can split a large slice of byte slices in smaller packets
type DataPacker interface {
	PackDataInChunks(data [][]byte, limit int) ([][]byte, error)
	IsInterfaceNil() bool
}

// TrieDataGetter returns requested data from the trie
type TrieDataGetter interface {
	GetSerializedNodes([]byte, uint64) ([][]byte, uint64, error)
	IsInterfaceNil() bool
}

// RequestedItemsHandler can determine if a certain key has or not been requested
type RequestedItemsHandler interface {
	Add(key string) error
	Has(key string) bool
	Sweep()
	IsInterfaceNil() bool
}

// P2PAntifloodHandler defines the behavior of a component able to signal that the system is too busy (or flooded) processing
// p2p messages
type P2PAntifloodHandler interface {
	CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
	CanProcessMessagesOnTopic(peer core.PeerID, topic string, numMessages uint32, _ uint64, sequence []byte) error
	BlacklistPeer(peer core.PeerID, reason string, duration time.Duration)
	IsInterfaceNil() bool
}

// WhiteListHandler is the interface needed to add whitelisted data
type WhiteListHandler interface {
	Remove(keys [][]byte)
	Add(keys [][]byte)
	IsInterfaceNil() bool
}

// ResolverDebugHandler defines an interface for debugging the reqested-resolved data
type ResolverDebugHandler interface {
	LogRequestedData(topic string, hashes [][]byte, numReqIntra int, numReqCross int)
	LogFailedToResolveData(topic string, hash []byte, err error)
	LogSucceededToResolveData(topic string, hash []byte)
	IsInterfaceNil() bool
}

// CurrentNetworkEpochProviderHandler is an interface needed to get the current epoch from the network
type CurrentNetworkEpochProviderHandler interface {
	SetNetworkEpochAtBootstrap(epoch uint32)
	EpochIsActiveInNetwork(epoch uint32) bool
	CurrentEpoch() uint32
	IsInterfaceNil() bool
}

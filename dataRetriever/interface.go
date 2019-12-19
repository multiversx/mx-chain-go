package dataRetriever

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// UnitType is the type for Storage unit identifiers
type UnitType uint8

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
	// MetaShardDataUnit is the metachain shard data unit identifier
	MetaShardDataUnit UnitType = 5
	// MetaPeerDataUnit is the metachain peer data unit identifier
	MetaPeerDataUnit UnitType = 6
	// UnsignedTransactionUnit is the unsigned transaction unit identifier
	UnsignedTransactionUnit UnitType = 7
	// RewardTransactionUnit is the reward transaction unit identifier
	RewardTransactionUnit UnitType = 8
	// MetaHdrNonceHashDataUnit is the meta header nonce-hash pair data unit identifier
	MetaHdrNonceHashDataUnit UnitType = 9
	// HeartbeatUnit is the heartbeat storage unit identifier
	HeartbeatUnit UnitType = 10
	// MiniBlockHeaderUnit is the miniblock header data unit identifier
	MiniBlockHeaderUnit = 11
	// BootstrapUnit is the bootstrap storage unit identifier
	BootstrapUnit UnitType = 11
	//StatusMetricsUnit is the status metrics storage unit identifier
	StatusMetricsUnit UnitType = 12

	// ShardHdrNonceHashDataUnit is the header nonce-hash pair data unit identifier
	//TODO: Add only unit types lower than 100
	ShardHdrNonceHashDataUnit UnitType = 100
	//TODO: Do not add unit type greater than 100 as the metachain creates this kind of unit type for each shard.
	//100 -> shard 0, 101 -> shard 1 and so on. This should be replaced with a factory which will manage the unit types
	//creation
)

// Resolver defines what a data resolver should do
type Resolver interface {
	RequestDataFromHash(hash []byte) error
	ProcessReceivedMessage(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error
	IsInterfaceNil() bool
}

// HeaderResolver defines what a block header resolver should do
type HeaderResolver interface {
	Resolver
	RequestDataFromNonce(nonce uint64) error
}

// MiniBlocksResolver defines what a mini blocks resolver should do
type MiniBlocksResolver interface {
	Resolver
	RequestDataFromHashArray(hashes [][]byte) error
	GetMiniBlocks(hashes [][]byte) (block.MiniBlockSlice, [][]byte)
	GetMiniBlocksFromPool(hashes [][]byte) (block.MiniBlockSlice, [][]byte)
}

// TopicResolverSender defines what sending operations are allowed for a topic resolver
type TopicResolverSender interface {
	SendOnRequestTopic(rd *RequestData) error
	Send(buff []byte, peer p2p.PeerID) error
	TopicRequestSuffix() string
	TargetShardID() uint32
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
	IsInterfaceNil() bool
}

// ResolversFinder extends a container resolver and have 2 additional functionality
type ResolversFinder interface {
	ResolversContainer
	IntraShardResolver(baseTopic string) (Resolver, error)
	MetaChainResolver(baseTopic string) (Resolver, error)
	CrossShardResolver(baseTopic string, crossShard uint32) (Resolver, error)
}

// ResolversContainerFactory defines the functionality to create a resolvers container
type ResolversContainerFactory interface {
	Create() (ResolversContainer, error)
	IsInterfaceNil() bool
}

// MessageHandler defines the functionality needed by structs to send data to other peers
type MessageHandler interface {
	ConnectedPeersOnTopic(topic string) []p2p.PeerID
	SendToConnectedPeer(topic string, buff []byte, peerID p2p.PeerID) error
	IsInterfaceNil() bool
}

// TopicHandler defines the functionality needed by structs to manage topics and message processors
type TopicHandler interface {
	HasTopic(name string) bool
	CreateTopic(name string, createChannelForTopic bool) error
	RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error
}

// TopicMessageHandler defines the functionality needed by structs to manage topics, message processors and to send data
// to other peers
type TopicMessageHandler interface {
	MessageHandler
	TopicHandler
}

// IntRandomizer interface provides functionality over generating integer numbers
type IntRandomizer interface {
	Intn(n int) (int, error)
	IsInterfaceNil() bool
}

// StorageType defines the storage levels on a node
type StorageType uint8

// DataRetriever interface provides functionality over high level data request component
type DataRetriever interface {
	// Get methods searches for data in storage units and returns results, it is a blocking function
	Get(keys [][]byte, identifier string, lowestLevel StorageType, haveTime func() time.Duration) (map[string]interface{}, [][]byte, error)
	// Has searches for a value identifier by a key in storage
	Has(key []byte, identifier string, level StorageType) (StorageType, error)
	// HasOrAdd searches and adds a value if not exist in storage
	HasOrAdd(key []byte, value interface{}, identifier string, level StorageType)
	// Remove deletes an element from storage level
	Remove(key []byte, identifier string, lowestLevel StorageType) error
	// Put saves a key-value pair into storage
	Put(key []byte, value interface{}, identifier string, level StorageType) error
	// Keys returns all the keys from an identifier and storage type
	Keys(identifier string, level StorageType)
	// Request searches for data in specified storage level, if not present launches threads to search in network
	Request(keys [][]byte, identifier string, level StorageType, haveTime func() time.Duration, callbackHandler func(key []byte)) (map[string]interface{}, [][]byte, error)
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// Notifier defines a way to register funcs that get called when something useful happens
type Notifier interface {
	RegisterHandler(func(key []byte))
	IsInterfaceNil() bool
}

// PeerListCreator is used to create a peer list
type PeerListCreator interface {
	PeerList() []p2p.PeerID
	IsInterfaceNil() bool
}

// ShardedDataCacherNotifier defines what a sharded-data structure can perform
type ShardedDataCacherNotifier interface {
	Notifier

	ShardDataStore(cacheId string) (c storage.Cacher)
	AddData(key []byte, data interface{}, cacheId string)
	SearchFirstData(key []byte) (value interface{}, ok bool)
	RemoveData(key []byte, cacheId string)
	RemoveSetOfDataFromPool(keys [][]byte, cacheId string)
	RemoveDataFromAllShards(key []byte)
	MergeShardStores(sourceCacheID, destCacheID string)
	MoveData(sourceCacheID, destCacheID string, key [][]byte)
	Clear()
	ClearShardStore(cacheId string)
	CreateShardStore(cacheId string)
}

// ShardIdHashMap represents a map for shardId and hash
type ShardIdHashMap interface {
	Load(shardId uint32) ([]byte, bool)
	Store(shardId uint32, hash []byte)
	Range(f func(shardId uint32, hash []byte) bool)
	Delete(shardId uint32)
	IsInterfaceNil() bool
}

// Uint64SyncMapCacher defines a cacher-type struct that uses uint64 keys and sync-maps values
type Uint64SyncMapCacher interface {
	Clear()
	Get(nonce uint64) (ShardIdHashMap, bool)
	Merge(nonce uint64, src ShardIdHashMap)
	Remove(nonce uint64, shardId uint32)
	RegisterHandler(handler func(nonce uint64, shardId uint32, value []byte))
	Has(nonce uint64, shardId uint32) bool
	IsInterfaceNil() bool
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
	Headers() storage.Cacher
	HeadersNonces() Uint64SyncMapCacher
	MiniBlocks() storage.Cacher
	PeerChangesBlocks() storage.Cacher
	MetaBlocks() storage.Cacher
	CurrentBlockTxs() TransactionCacher
	IsInterfaceNil() bool
}

// MetaPoolsHolder defines getter for data pools for metachain
type MetaPoolsHolder interface {
	MetaBlocks() storage.Cacher
	MiniBlocks() storage.Cacher
	ShardHeaders() storage.Cacher
	HeadersNonces() Uint64SyncMapCacher
	Transactions() ShardedDataCacherNotifier
	UnsignedTransactions() ShardedDataCacherNotifier
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
	// GetAll gets all the elements with keys in the keys array, from the selected storage unit
	// If there is a missing key in the unit, it returns an error
	GetAll(unitType UnitType, keys [][]byte) (map[string][]byte, error)
	// Destroy removes the underlying files/resources used by the storage service
	Destroy() error
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// DataPacker can split a large slice of byte slices in smaller packets
type DataPacker interface {
	PackDataInChunks(data [][]byte, limit int) ([][]byte, error)
	IsInterfaceNil() bool
}

// RequestedItemsHandler can determine if a certain key has or not been requested
type RequestedItemsHandler interface {
	Add(key string) error
	Has(key string) bool
	Sweep()
	IsInterfaceNil() bool
}

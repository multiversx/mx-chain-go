package dataRetriever

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// Resolver defines what a data resolver should do
type Resolver interface {
	RequestDataFromHash(hash []byte) error
	ProcessReceivedMessage(message p2p.MessageP2P) error
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

	// TODO miniblockresolver should not know about miniblockslice
	GetMiniBlocks(hashes [][]byte) block.MiniBlockSlice
}

// TopicResolverSender defines what sending operations are allowed for a topic resolver
type TopicResolverSender interface {
	SendOnRequestTopic(rd *RequestData) error
	Send(buff []byte, peer p2p.PeerID) error
	TopicRequestSuffix() string
}

// ResolversContainer defines a resolvers holder data type with basic functionality
type ResolversContainer interface {
	Get(key string) (Resolver, error)
	Add(key string, val Resolver) error
	AddMultiple(keys []string, resolvers []Resolver) error
	Replace(key string, val Resolver) error
	Remove(key string)
	Len() int
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
}

// MessageHandler defines the functionality needed by structs to send data to other peers
type MessageHandler interface {
	ConnectedPeersOnTopic(topic string) []p2p.PeerID
	SendToConnectedPeer(topic string, buff []byte, peerID p2p.PeerID) error
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
}

type StorageType uint8

const (
	CACHE     StorageType = 0
	LOCALDISC StorageType = 1
	NETWORK   StorageType = 2
)

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
	// MetaShardDataUnit is the metachain shard data unit identifier
	MetaShardDataUnit UnitType = 5
	// MetaPeerDataUnit is the metachain peer data unit identifier
	MetaPeerDataUnit UnitType = 6
	//
)

// UnitType is the type for Storage unit identifiers
type UnitType uint8

// Notifier defines a way to register funcs that get called when something useful happens
type Notifier interface {
	RegisterHandler(func(key []byte))
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

// Uint64Cacher defines a cacher-type struct that uses uint64 keys and []byte values (usually hashes)
type Uint64Cacher interface {
	Clear()
	Put(uint64, interface{}) bool
	Get(uint64) (interface{}, bool)
	Has(uint64) bool
	Peek(uint64) (interface{}, bool)
	HasOrAdd(uint64, interface{}) (bool, bool)
	Remove(uint64)
	RemoveOldest()
	Keys() []uint64
	Len() int
	RegisterHandler(handler func(nonce uint64))
}

// PoolsHolder defines getters for data pools
type PoolsHolder interface {
	Transactions() ShardedDataCacherNotifier
	Headers() storage.Cacher
	HeadersNonces() Uint64Cacher
	MiniBlocks() storage.Cacher
	PeerChangesBlocks() storage.Cacher
	MetaBlocks() storage.Cacher
	MetaHeadersNonces() Uint64Cacher
}

// MetaPoolsHolder defines getter for data pools for metachain
type MetaPoolsHolder interface {
	MetaChainBlocks() storage.Cacher
	MiniBlockHashes() ShardedDataCacherNotifier
	ShardHeaders() storage.Cacher
	MetaBlockNonces() Uint64Cacher
	ShardHeadersNonces() Uint64Cacher
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
}

// DataPacker can split a large slice of byte slices in smaller packets
type DataPacker interface {
	PackDataInChunks(data [][]byte, limit int) ([][]byte, error)
}

package dataRetriever

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/counting"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
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

// PeerAuthenticationResolver defines what a peer authentication resolver should do
type PeerAuthenticationResolver interface {
	Resolver
	RequestDataFromHashArray(hashes [][]byte, epoch uint32) error
}

// ValidatorInfoResolver defines what a validator info resolver should do
type ValidatorInfoResolver interface {
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
	CrossShardPeerList() []core.PeerID
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
	Keys() [][]byte
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

// TransactionCacher defines the methods for the local transaction cacher, needed for the current block
type TransactionCacher interface {
	Clean()
	GetTx(txHash []byte) (data.TransactionHandler, error)
	AddTx(txHash []byte, tx data.TransactionHandler)
	IsInterfaceNil() bool
}

// ValidatorInfoCacher defines the methods for the local validator info cacher, needed for the current epoch
type ValidatorInfoCacher interface {
	Clean()
	GetValidatorInfo(validatorInfoHash []byte) (*state.ShardValidatorInfo, error)
	AddValidatorInfo(validatorInfoHash []byte, validatorInfo *state.ShardValidatorInfo)
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
	TrieNodesChunks() storage.Cacher
	SmartContracts() storage.Cacher
	CurrentBlockTxs() TransactionCacher
	CurrentEpochValidatorInfo() ValidatorInfoCacher
	PeerAuthentications() storage.Cacher
	Heartbeats() storage.Cacher
	ValidatorsInfo() ShardedDataCacherNotifier
	Close() error
	IsInterfaceNil() bool
}

// StorageService is the interface for data storage unit provided services
type StorageService interface {
	// GetStorer returns the storer from the chain map
	// If the unit is missing, it returns an error
	GetStorer(unitType UnitType) (storage.Storer, error)
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
	// GetAllStorers returns all the storers
	GetAllStorers() map[UnitType]storage.Storer
	// Destroy removes the underlying files/resources used by the storage service
	Destroy() error
	// CloseAll will close all the units
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
	GetSerializedNode([]byte) ([]byte, error)
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

// CurrentNetworkEpochProviderHandler is an interface able to compute if the provided epoch is active on the network or not
type CurrentNetworkEpochProviderHandler interface {
	EpochIsActiveInNetwork(epoch uint32) bool
	EpochConfirmed(newEpoch uint32, newTimestamp uint64)
	IsInterfaceNil() bool
}

// PreferredPeersHolderHandler defines the behavior of a component able to handle preferred peers operations
type PreferredPeersHolderHandler interface {
	Get() map[uint32][]core.PeerID
	Contains(peerID core.PeerID) bool
	IsInterfaceNil() bool
}

// PeersRatingHandler represent an entity able to handle peers ratings
type PeersRatingHandler interface {
	AddPeer(pid core.PeerID)
	IncreaseRating(pid core.PeerID)
	DecreaseRating(pid core.PeerID)
	GetTopRatedPeersFromList(peers []core.PeerID, minNumOfPeersExpected int) []core.PeerID
	IsInterfaceNil() bool
}

// SelfShardIDProvider defines the behavior of a component able to provide the self shard ID
type SelfShardIDProvider interface {
	SelfId() uint32
	IsInterfaceNil() bool
}

// NodesCoordinator provides Validator methods needed for the peer processing
type NodesCoordinator interface {
	GetAllEligibleValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error)
	IsInterfaceNil() bool
}

// PeerAuthenticationPayloadValidator defines the operations supported by an entity able to validate timestamps
// found in peer authentication messages
type PeerAuthenticationPayloadValidator interface {
	ValidateTimestamp(payloadTimestamp int64) error
	IsInterfaceNil() bool
}

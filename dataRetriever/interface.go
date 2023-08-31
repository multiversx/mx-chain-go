package dataRetriever

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/counting"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
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
	ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) error
	SetDebugHandler(handler DebugHandler) error
	Close() error
	IsInterfaceNil() bool
}

// Requester defines what a data requester should do
type Requester interface {
	RequestDataFromHash(hash []byte, epoch uint32) error
	SetNumPeersToQuery(intra int, cross int)
	NumPeersToQuery() (int, int)
	SetDebugHandler(handler DebugHandler) error
	IsInterfaceNil() bool
}

// HeaderResolver defines what a block header resolver should do
type HeaderResolver interface {
	Resolver
	SetEpochHandler(epochHandler EpochHandler) error
}

// HeaderRequester defines what a block header requester should do
type HeaderRequester interface {
	Requester
	SetEpochHandler(epochHandler EpochHandler) error
}

// TopicResolverSender defines what sending operations are allowed for a topic resolver
type TopicResolverSender interface {
	Send(buff []byte, peer core.PeerID, destination p2p.MessageHandler) error
	RequestTopic() string
	TargetShardID() uint32
	SetDebugHandler(handler DebugHandler) error
	DebugHandler() DebugHandler
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

// RequestersFinder extends a requesters container
type RequestersFinder interface {
	RequestersContainer
	IntraShardRequester(baseTopic string) (Requester, error)
	MetaChainRequester(baseTopic string) (Requester, error)
	CrossShardRequester(baseTopic string, crossShard uint32) (Requester, error)
	MetaCrossShardRequester(baseTopic string, crossShard uint32) (Requester, error)
}

// ResolversContainerFactory defines the functionality to create a resolvers container
type ResolversContainerFactory interface {
	Create() (ResolversContainer, error)
	IsInterfaceNil() bool
}

// TopicRequestSender defines what sending operations are allowed for a topic requester
type TopicRequestSender interface {
	SendOnRequestTopic(rd *RequestData, originalHashes [][]byte) error
	SetNumPeersToQuery(intra int, cross int)
	NumPeersToQuery() (int, int)
	RequestTopic() string
	TargetShardID() uint32
	SetDebugHandler(handler DebugHandler) error
	DebugHandler() DebugHandler
	IsInterfaceNil() bool
}

// RequestersContainer defines a requesters holder data type with basic functionality
type RequestersContainer interface {
	Get(key string) (Requester, error)
	Add(key string, val Requester) error
	AddMultiple(keys []string, requesters []Requester) error
	Replace(key string, val Requester) error
	Remove(key string)
	Len() int
	RequesterKeys() string
	Iterate(handler func(key string, requester Requester) bool)
	Close() error
	IsInterfaceNil() bool
}

// RequestersContainerFactory defines the functionality to create a requesters container
type RequestersContainerFactory interface {
	Create() (RequestersContainer, error)
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
	SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error
	ID() core.PeerID
	ConnectedPeers() []core.PeerID
	IsConnected(peerID core.PeerID) bool
	IsInterfaceNil() bool
}

// TopicHandler defines the functionality needed by structs to manage topics and message processors
type TopicHandler interface {
	HasTopic(name string) bool
	CreateTopic(name string, createChannelForTopic bool) error
	RegisterMessageProcessor(topic string, identifier string, handler p2p.MessageProcessor) error
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
	UserTransactions() ShardedDataCacherNotifier
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

// DebugHandler defines an interface for debugging the requested-resolved data
type DebugHandler interface {
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

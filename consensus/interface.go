package consensus

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
)

// BlsConsensusType specifies the signature scheme used in the consensus
const BlsConsensusType = "bls"

// RoundHandler defines the actions which should be handled by a round implementation
type RoundHandler interface {
	Index() int64
	BeforeGenesis() bool
	// UpdateRound updates the index and the time stamp of the round depending of the genesis time and the current time given
	UpdateRound(time.Time, time.Time)
	TimeStamp() time.Time
	TimeDuration() time.Duration
	RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration
	IsInterfaceNil() bool
}

// SubroundHandler defines the actions which should be handled by a subround implementation
type SubroundHandler interface {
	// DoWork implements of the subround's job
	DoWork(ctx context.Context, roundHandler RoundHandler) bool
	// Previous returns the ID of the previous subround
	Previous() int
	// Next returns the ID of the next subround
	Next() int
	// Current returns the ID of the current subround
	Current() int
	// StartTime returns the start time, in the roundHandler time, of the current subround
	StartTime() int64
	// EndTime returns the top limit time, in the roundHandler time, of the current subround
	EndTime() int64
	// Name returns the name of the current roundHandler
	Name() string
	// ConsensusChannel returns the consensus channel
	ConsensusChannel() chan bool
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// ChronologyHandler defines the actions which should be handled by a chronology implementation
type ChronologyHandler interface {
	Close() error
	AddSubround(SubroundHandler)
	RemoveAllSubrounds()
	// StartRounds starts rounds in a sequential manner, one after the other
	StartRounds()
	IsInterfaceNil() bool
}

// BroadcastMessenger defines the behaviour of the broadcast messages by the consensus group
type BroadcastMessenger interface {
	BroadcastBlock(data.BodyHandler, data.HeaderHandler) error
	BroadcastHeader(data.HeaderHandler) error
	BroadcastMiniBlocks(map[uint32][]byte) error
	BroadcastTransactions(map[string][][]byte) error
	BroadcastConsensusMessage(*Message) error
	BroadcastBlockDataLeader(header data.HeaderHandler, miniBlocks map[uint32][]byte, transactions map[string][][]byte) error
	PrepareBroadcastHeaderValidator(header data.HeaderHandler, miniBlocks map[uint32][]byte, transactions map[string][][]byte, order int)
	PrepareBroadcastBlockDataValidator(header data.HeaderHandler, miniBlocks map[uint32][]byte, transactions map[string][][]byte, idx int)
	IsInterfaceNil() bool
}

// P2PMessenger defines a subset of the p2p.Messenger interface
type P2PMessenger interface {
	Broadcast(topic string, buff []byte)
	IsInterfaceNil() bool
}

// NetworkShardingCollector defines the updating methods used by the network sharding component
// The interface assures that the collected data will be used by the p2p network sharding components
type NetworkShardingCollector interface {
	UpdatePeerIDInfo(pid core.PeerID, pk []byte, shardID uint32)
	PutPeerIdSubType(pid core.PeerID, peerSubType core.P2PPeerSubType)
	GetPeerInfo(pid core.PeerID) core.P2PPeerInfo
	IsInterfaceNil() bool
}

// P2PAntifloodHandler defines the behavior of a component able to signal that the system is too busy (or flooded) processing
// p2p messages
type P2PAntifloodHandler interface {
	CanProcessMessage(message core.MessageP2P, fromConnectedPeer core.PeerID) error
	CanProcessMessagesOnTopic(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error
	ResetForTopic(topic string)
	SetMaxMessagesForTopic(topic string, maxNum uint32)
	BlacklistPeer(peer core.PeerID, reason string, duration time.Duration)
	IsInterfaceNil() bool
}

// HeadersPoolSubscriber can subscribe for notifications when a new block header is added to the headers pool
type HeadersPoolSubscriber interface {
	RegisterHandler(handler func(headerHandler data.HeaderHandler, headerHash []byte))
	IsInterfaceNil() bool
}

// PeerHonestyHandler defines the behaivour of a component able to handle/monitor the peer honesty of nodes which are
// participating in consensus
type PeerHonestyHandler interface {
	ChangeScore(pk string, topic string, units int)
	IsInterfaceNil() bool
	Close() error
}

// InterceptorSubscriber can subscribe for notifications when data is received by an interceptor
type InterceptorSubscriber interface {
	RegisterHandler(handler func(toShard uint32, data []byte))
	IsInterfaceNil() bool
}

// HeaderSigVerifier encapsulates methods that are check if header rand seed, leader signature and aggregate signature are correct
type HeaderSigVerifier interface {
	VerifyRandSeed(header data.HeaderHandler) error
	VerifyLeaderSignature(header data.HeaderHandler) error
	VerifySignature(header data.HeaderHandler) error
	IsInterfaceNil() bool
}

// FallbackHeaderValidator defines the behaviour of a component able to signal when a fallback header validation could be applied
type FallbackHeaderValidator interface {
	ShouldApplyFallbackValidation(headerHandler data.HeaderHandler) bool
	IsInterfaceNil() bool
}

// NodeRedundancyHandler provides functionality to handle the redundancy mechanism for a node
type NodeRedundancyHandler interface {
	IsRedundancyNode() bool
	IsMainMachineActive() bool
	AdjustInactivityIfNeeded(selfPubKey string, consensusPubKeys []string, roundIndex int64)
	ResetInactivityIfNeeded(selfPubKey string, consensusMsgPubKey string, consensusMsgPeerID core.PeerID)
	ObserverPrivateKey() crypto.PrivateKey
	IsInterfaceNil() bool
}

// ScheduledProcessor encapsulates the scheduled processor functionality required by consensus module
type ScheduledProcessor interface {
	StartScheduledProcessing(header data.HeaderHandler, body data.BodyHandler, startTime time.Time)
	ForceStopScheduledExecutionBlocking()
	IsProcessedOKWithTimeout() bool
	IsInterfaceNil() bool
}

// SignatureHandler defines the behaviour of a component that handles signatures in consensus
type SignatureHandler interface {
	Reset(pubKeys []string) error
	CreateSignatureShare(msg []byte, index uint16, epoch uint32) ([]byte, error)
	StoreSignatureShare(index uint16, sig []byte) error
	SignatureShare(index uint16) ([]byte, error)
	VerifySignatureShare(index uint16, sig []byte, msg []byte, epoch uint32) error
	AggregateSigs(bitmap []byte, epoch uint32) ([]byte, error)
	SetAggregatedSig([]byte) error
	Verify(msg []byte, bitmap []byte, epoch uint32) error
	IsInterfaceNil() bool
}

// InterceptorsContainer defines an interceptors holder data type with basic functionality
type InterceptorsContainer interface {
	Get(key string) (core.Interceptor, error)
	Add(key string, val core.Interceptor) error
	AddMultiple(keys []string, interceptors []core.Interceptor) error
	Replace(key string, val core.Interceptor) error
	Remove(key string)
	Len() int
	Iterate(handler func(key string, interceptor core.Interceptor) bool)
	Close() error
	IsInterfaceNil() bool
}

// ShardCoordinator defines what a shard state coordinator should hold
type ShardCoordinator interface {
	NumberOfShards() uint32
	ComputeId(address []byte) uint32
	SelfId() uint32
	SameShard(firstAddress, secondAddress []byte) bool
	CommunicationIdentifier(destShardID uint32) string
	IsInterfaceNil() bool
}

// Cacher provides caching services
type Cacher interface {
	Put(key []byte, value interface{}, sizeInBytes int) (evicted bool)
	Get(key []byte) (value interface{}, ok bool)
	IsInterfaceNil() bool
}

// OutportHandler is interface that defines what a proxy implementation should be able to do
// The node is able to talk only with this interface
type OutportHandler interface {
	SaveRoundsInfo(roundsInfos []*indexer.RoundInfo)
	HasDrivers() bool
	IsInterfaceNil() bool
}

// SyncTimer defines an interface for time synchronization
type SyncTimer interface {
	CurrentTime() time.Time
	FormattedCurrentTime() string
	IsInterfaceNil() bool
}

// RegistrationHandler provides Register and Unregister functionality for the end of epoch events
type RegistrationHandler interface {
	RegisterHandler(handler core.EpochStartActionHandler)
	UnregisterHandler(handler core.EpochStartActionHandler)
	IsInterfaceNil() bool
}

// BlockProcessor is the main interface for block execution engine
type BlockProcessor interface {
	ProcessBlock(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	ProcessScheduledBlock(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlock(header data.HeaderHandler, body data.BodyHandler) error
	CreateBlock(initialHdr data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error)
	CreateNewHeader(round uint64, nonce uint64) (data.HeaderHandler, error)
	MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error)
	DecodeBlockBody(dta []byte) data.BodyHandler
	DecodeBlockHeader(dta []byte) data.HeaderHandler
	RevertCurrentBlock()
	IsInterfaceNil() bool
}

// Bootstrapper is an interface that defines the behaviour of a struct that is able
// to synchronize the node
type Bootstrapper interface {
	Close() error
	AddSyncStateListener(func(isSyncing bool))
	GetNodeState() core.NodeState
	StartSyncingBlocks()
	IsInterfaceNil() bool
}

// NodesCoordinator defines the behaviour of a struct able to do validator group selection
type NodesCoordinator interface {
	GetValidatorsIndexes(publicKeys []string, epoch uint32) ([]uint64, error)
	ComputeConsensusGroup(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []core.Validator, err error)
	ShardIdForEpoch(epoch uint32) (uint32, error)
	GetConsensusWhitelistedNodes(epoch uint32) (map[string]struct{}, error)
	ConsensusGroupSize(uint32) int
	IsInterfaceNil() bool
}

// ScheduledBlockProcessor is the interface for the scheduled miniBlocks execution part of the block processor
type ScheduledBlockProcessor interface {
	ProcessScheduledBlock(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	IsInterfaceNil() bool
}

// RoundTimeDurationHandler defines the methods to get the time duration of a round
type RoundTimeDurationHandler interface {
	TimeDuration() time.Duration
	IsInterfaceNil() bool
}

// ForkDetector is an interface that defines the behaviour of a struct that is able
// to detect forks
type ForkDetector interface {
	AddHeader(header data.HeaderHandler, headerHash []byte, state core.BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error
	IsInterfaceNil() bool
}

// HeaderIntegrityVerifier encapsulates methods useful to check that a header's integrity is correct
type HeaderIntegrityVerifier interface {
	Verify(header data.HeaderHandler) error
	GetVersion(epoch uint32) string
	IsInterfaceNil() bool
}

// MultiSignerContainer defines the container for different versioned multiSigner instances
type MultiSignerContainer interface {
	GetMultiSigner(epoch uint32) (crypto.MultiSigner, error)
	IsInterfaceNil() bool
}

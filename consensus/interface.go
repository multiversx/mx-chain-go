package consensus

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/p2p"
)

// BlsConsensusType specifies the signature scheme used in the consensus
const BlsConsensusType = "bls"

// RoundHandler defines the actions which should be handled by a round implementation
type RoundHandler interface {
	Index() int64
	BeforeGenesis() bool
	// UpdateRound updates the index and the time stamp of the round depending on the genesis time and the current time given
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

// SubRoundStartExtraSignatureHandler defines an extra signer during start subround in a consensus process
type SubRoundStartExtraSignatureHandler interface {
	Reset(pubKeys []string) error
	Identifier() string
	IsInterfaceNil() bool
}

// SubRoundSignatureExtraSignatureHandler defines an extra signer during signature subround in a consensus process
type SubRoundSignatureExtraSignatureHandler interface {
	CreateSignatureShare(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) ([]byte, error)
	AddSigShareToConsensusMessage(sigShare []byte, cnsMsg *Message) error
	StoreSignatureShare(index uint16, cnsMsg *Message) error
	Identifier() string
	IsInterfaceNil() bool
}

// SubRoundEndExtraSignatureHandler defines an extra signer during end subround in a consensus process
type SubRoundEndExtraSignatureHandler interface {
	AggregateAndSetSignatures(bitmap []byte, header data.HeaderHandler) ([]byte, error)
	AddLeaderAndAggregatedSignatures(header data.HeaderHandler, cnsMsg *Message) error
	SignAndSetLeaderSignature(header data.HeaderHandler, leaderPubKey []byte) error
	SetAggregatedSignatureInHeader(header data.HeaderHandler, aggregatedSig []byte) error
	SetConsensusDataInHeader(header data.HeaderHandler, cnsMsg *Message) error
	VerifyAggregatedSignatures(bitmap []byte, header data.HeaderHandler) error
	Identifier() string
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
	BroadcastHeader(data.HeaderHandler, []byte) error
	BroadcastMiniBlocks(map[uint32][]byte, []byte) error
	BroadcastTransactions(map[string][][]byte, []byte) error
	BroadcastConsensusMessage(*Message) error
	BroadcastBlockDataLeader(header data.HeaderHandler, miniBlocks map[uint32][]byte, transactions map[string][][]byte, pkBytes []byte) error
	PrepareBroadcastHeaderValidator(header data.HeaderHandler, miniBlocks map[uint32][]byte, transactions map[string][][]byte, idx int, pkBytes []byte)
	PrepareBroadcastBlockDataValidator(header data.HeaderHandler, miniBlocks map[uint32][]byte, transactions map[string][][]byte, idx int, pkBytes []byte)
	IsInterfaceNil() bool
}

// P2PMessenger defines a subset of the p2p.Messenger interface
type P2PMessenger interface {
	Broadcast(topic string, buff []byte)
	BroadcastUsingPrivateKey(topic string, buff []byte, pid core.PeerID, skBytes []byte)
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
	CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
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

// P2PSigningHandler defines the behaviour of a component able to verify p2p message signature
type P2PSigningHandler interface {
	Verify(message p2p.MessageP2P) error
	Serialize(messages []p2p.MessageP2P) ([]byte, error)
	Deserialize(messagesBytes []byte) ([]p2p.MessageP2P, error)
	IsInterfaceNil() bool
}

// PeerBlacklistHandler defines the behaviour of a component able to blacklist p2p peers
type PeerBlacklistHandler interface {
	IsPeerBlacklisted(peer core.PeerID) bool
	BlacklistPeer(peer core.PeerID, duration time.Duration)
	Close() error
	IsInterfaceNil() bool
}

// SigningHandler defines the behaviour of a component that handles multi and single signatures used in consensus operations
type SigningHandler interface {
	Reset(pubKeys []string) error
	CreateSignatureShareForPublicKey(message []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error)
	CreateSignatureForPublicKey(message []byte, publicKeyBytes []byte) ([]byte, error)
	VerifySingleSignature(publicKeyBytes []byte, message []byte, signature []byte) error
	StoreSignatureShare(index uint16, sig []byte) error
	SignatureShare(index uint16) ([]byte, error)
	VerifySignatureShare(index uint16, sig []byte, msg []byte, epoch uint32) error
	AggregateSigs(bitmap []byte, epoch uint32) ([]byte, error)
	SetAggregatedSig([]byte) error
	Verify(msg []byte, bitmap []byte, epoch uint32) error
	ShallowClone() SigningHandler
	IsInterfaceNil() bool
}

// KeysHandler defines the operations implemented by a component that will manage all keys,
// including the single signer keys or the set of multi-keys
type KeysHandler interface {
	GetHandledPrivateKey(pkBytes []byte) crypto.PrivateKey
	GetP2PIdentity(pkBytes []byte) ([]byte, core.PeerID, error)
	IsKeyManagedByCurrentNode(pkBytes []byte) bool
	IncrementRoundsWithoutReceivedMessages(pkBytes []byte)
	GetAssociatedPid(pkBytes []byte) core.PeerID
	IsOriginalPublicKeyOfTheNode(pkBytes []byte) bool
	ResetRoundsWithoutReceivedMessages(pkBytes []byte, pid core.PeerID)
	IsInterfaceNil() bool
}

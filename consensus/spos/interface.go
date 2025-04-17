package spos

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-go/common"
	cryptoCommon "github.com/multiversx/mx-chain-go/common/crypto"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/ntp"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

// ConsensusCoreHandler encapsulates all needed data for the Consensus
type ConsensusCoreHandler interface {
	Blockchain() data.ChainHandler
	BlockProcessor() process.BlockProcessor
	BootStrapper() process.Bootstrapper
	BroadcastMessenger() consensus.BroadcastMessenger
	Chronology() consensus.ChronologyHandler
	GetAntiFloodHandler() consensus.P2PAntifloodHandler
	Hasher() hashing.Hasher
	Marshalizer() marshal.Marshalizer
	MultiSignerContainer() cryptoCommon.MultiSignerContainer
	RoundHandler() consensus.RoundHandler
	ShardCoordinator() sharding.Coordinator
	SyncTimer() ntp.SyncTimer
	NodesCoordinator() nodesCoordinator.NodesCoordinator
	EpochStartRegistrationHandler() epochStart.RegistrationHandler
	PeerHonestyHandler() consensus.PeerHonestyHandler
	HeaderSigVerifier() consensus.HeaderSigVerifier
	FallbackHeaderValidator() consensus.FallbackHeaderValidator
	NodeRedundancyHandler() consensus.NodeRedundancyHandler
	ScheduledProcessor() consensus.ScheduledProcessor
	MessageSigningHandler() consensus.P2PSigningHandler
	PeerBlacklistHandler() consensus.PeerBlacklistHandler
	SigningHandler() consensus.SigningHandler
	EnableEpochsHandler() common.EnableEpochsHandler
	EquivalentProofsPool() consensus.EquivalentProofsPool
	EpochNotifier() process.EpochNotifier
	InvalidSignersCache() InvalidSignersCache
	IsInterfaceNil() bool
}

// ConsensusService encapsulates the methods specifically for a consensus type (bls, bn)
// and will be used in the sposWorker
type ConsensusService interface {
	// InitReceivedMessages initializes the MessagesType map for all messages for the current ConsensusService
	InitReceivedMessages() map[consensus.MessageType][]*consensus.Message
	// GetStringValue gets the name of the messageType
	GetStringValue(consensus.MessageType) string
	// GetSubroundName gets the subround name for the subround id provided
	GetSubroundName(int) string
	// GetMessageRange provides the MessageType range used in checks by the consensus
	GetMessageRange() []consensus.MessageType
	// CanProceed returns if the current messageType can proceed further if previous subrounds finished
	CanProceed(*ConsensusState, consensus.MessageType) bool
	// IsMessageWithBlockBodyAndHeader returns if the current messageType is about block body and header
	IsMessageWithBlockBodyAndHeader(consensus.MessageType) bool
	// IsMessageWithBlockBody returns if the current messageType is about block body
	IsMessageWithBlockBody(consensus.MessageType) bool
	// IsMessageWithBlockHeader returns if the current messageType is about block header
	IsMessageWithBlockHeader(consensus.MessageType) bool
	// IsMessageWithSignature returns if the current messageType is about signature
	IsMessageWithSignature(consensus.MessageType) bool
	// IsMessageWithFinalInfo returns if the current messageType is about header final info
	IsMessageWithFinalInfo(consensus.MessageType) bool
	// IsMessageTypeValid returns if the current messageType is valid
	IsMessageTypeValid(consensus.MessageType) bool
	// IsMessageWithInvalidSigners returns if the current messageType is with invalid signers
	IsMessageWithInvalidSigners(consensus.MessageType) bool
	// IsSubroundSignature returns if the current subround is about signature
	IsSubroundSignature(int) bool
	// IsSubroundStartRound returns if the current subround is about start round
	IsSubroundStartRound(int) bool
	// GetMaxMessagesInARoundPerPeer returns the maximum number of messages a peer can send per round
	GetMaxMessagesInARoundPerPeer() uint32
	// GetMaxNumOfMessageTypeAccepted returns the maximum number of accepted consensus message types per round, per public key
	GetMaxNumOfMessageTypeAccepted(msgType consensus.MessageType) uint32
	// GetMessageTypeBlockHeader returns the message type for the block header
	GetMessageTypeBlockHeader() consensus.MessageType
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// WorkerHandler represents the interface for the SposWorker
type WorkerHandler interface {
	Close() error
	StartWorking()
	// AddReceivedMessageCall adds a new handler function for a received message type
	AddReceivedMessageCall(messageType consensus.MessageType, receivedMessageCall func(ctx context.Context, cnsDta *consensus.Message) bool)
	// AddReceivedHeaderHandler adds a new handler function for a received header
	AddReceivedHeaderHandler(handler func(data.HeaderHandler))
	// RemoveAllReceivedHeaderHandlers removes all the functions handlers
	RemoveAllReceivedHeaderHandlers()
	// AddReceivedProofHandler adds a new handler function for a received proof
	AddReceivedProofHandler(handler func(consensus.ProofHandler))
	// RemoveAllReceivedMessagesCalls removes all the functions handlers
	RemoveAllReceivedMessagesCalls()
	// ProcessReceivedMessage method redirects the received message to the channel which should handle it
	ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) ([]byte, error)
	// Extend does an extension for the subround with subroundId
	Extend(subroundId int)
	// GetConsensusStateChangedChannel gets the channel for the consensusStateChanged
	GetConsensusStateChangedChannel() chan bool
	// ExecuteStoredMessages tries to execute all the messages received which are valid for execution
	ExecuteStoredMessages()
	// DisplayStatistics method displays statistics of worker at the end of the round
	DisplayStatistics()
	// ReceivedHeader method is a wired method through which worker will receive headers from network
	ReceivedHeader(headerHandler data.HeaderHandler, headerHash []byte)
	// ResetConsensusMessages resets at the start of each round all the previous consensus messages received and equivalent messages, keeping the provided proofs
	ResetConsensusMessages()
	// ResetConsensusRoundState resets the consensus round state when transitioning to a different consensus version
	ResetConsensusRoundState()
	// ResetInvalidSignersCache resets the invalid signers cache
	ResetInvalidSignersCache()
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// PoolAdder adds data in a key-value pool
type PoolAdder interface {
	Put(key []byte, value interface{}, sizeInBytes int) (evicted bool)
	IsInterfaceNil() bool
}

// HeaderSigVerifier encapsulates methods that check if header signature is correct
type HeaderSigVerifier interface {
	VerifyRandSeed(header data.HeaderHandler) error
	VerifyLeaderSignature(header data.HeaderHandler) error
	VerifySignature(header data.HeaderHandler) error
	VerifySignatureForHash(header data.HeaderHandler, hash []byte, pubkeysBitmap []byte, signature []byte) error
	VerifyHeaderProof(headerProof data.HeaderProofHandler) error
	IsInterfaceNil() bool
}

// ConsensusDataIndexer defines the actions that a consensus data indexer has to do
type ConsensusDataIndexer interface {
	SaveRoundsInfo(roundsInfos []*outport.RoundInfo)
	IsInterfaceNil() bool
}

// PeerBlackListCacher can determine if a certain peer id is blacklisted or not
type PeerBlackListCacher interface {
	Upsert(pid core.PeerID, span time.Duration) error
	Has(pid core.PeerID) bool
	Sweep()
	IsInterfaceNil() bool
}

// SentSignaturesTracker defines a component able to handle sent signature from self
type SentSignaturesTracker interface {
	StartRound()
	SignatureSent(pkBytes []byte)
	IsInterfaceNil() bool
}

// ConsensusStateHandler encapsulates all needed data for the Consensus
type ConsensusStateHandler interface {
	ResetConsensusState()
	ResetConsensusRoundState()
	AddReceivedHeader(headerHandler data.HeaderHandler)
	GetReceivedHeaders() []data.HeaderHandler
	AddMessageWithSignature(key string, message p2p.MessageP2P)
	GetMessageWithSignature(key string) (p2p.MessageP2P, bool)
	IsNodeLeaderInCurrentRound(node string) bool
	GetLeader() (string, error)
	GetNextConsensusGroup(
		randomSource []byte,
		round uint64,
		shardId uint32,
		nodesCoordinator nodesCoordinator.NodesCoordinator,
		epoch uint32,
	) (string, []string, error)
	IsConsensusDataSet() bool
	IsConsensusDataEqual(data []byte) bool
	IsJobDone(node string, currentSubroundId int) bool
	IsSubroundFinished(subroundID int) bool
	IsNodeSelf(node string) bool
	IsBlockBodyAlreadyReceived() bool
	IsHeaderAlreadyReceived() bool
	CanDoSubroundJob(currentSubroundId int) bool
	CanProcessReceivedMessage(cnsDta *consensus.Message, currentRoundIndex int64, currentSubroundId int) bool
	GenerateBitmap(subroundId int) []byte
	ProcessingBlock() bool
	SetProcessingBlock(processingBlock bool)
	GetData() []byte
	SetData(data []byte)
	IsMultiKeyLeaderInCurrentRound() bool
	IsLeaderJobDone(currentSubroundId int) bool
	IsMultiKeyJobDone(currentSubroundId int) bool
	IsSelfJobDone(currentSubroundID int) bool
	GetMultikeyRedundancyStepInReason() string
	ResetRoundsWithoutReceivedMessages(pkBytes []byte, pid core.PeerID)
	GetRoundCanceled() bool
	SetRoundCanceled(state bool)
	GetRoundIndex() int64
	SetRoundIndex(roundIndex int64)
	GetRoundTimeStamp() time.Time
	SetRoundTimeStamp(roundTimeStamp time.Time)
	GetExtendedCalled() bool
	GetBody() data.BodyHandler
	SetBody(body data.BodyHandler)
	GetHeader() data.HeaderHandler
	SetHeader(header data.HeaderHandler)
	GetWaitingAllSignaturesTimeOut() bool
	SetWaitingAllSignaturesTimeOut(bool)
	RoundConsensusHandler
	RoundStatusHandler
	RoundThresholdHandler
	IsInterfaceNil() bool
}

// RoundConsensusHandler encapsulates the methods needed for a consensus round
type RoundConsensusHandler interface {
	ConsensusGroupIndex(pubKey string) (int, error)
	SelfConsensusGroupIndex() (int, error)
	SetEligibleList(eligibleList map[string]struct{})
	ConsensusGroup() []string
	SetConsensusGroup(consensusGroup []string)
	SetLeader(leader string)
	ConsensusGroupSize() int
	SetConsensusGroupSize(consensusGroupSize int)
	SelfPubKey() string
	SetSelfPubKey(selfPubKey string)
	JobDone(key string, subroundId int) (bool, error)
	SetJobDone(key string, subroundId int, value bool) error
	SelfJobDone(subroundId int) (bool, error)
	IsNodeInConsensusGroup(node string) bool
	IsNodeInEligibleList(node string) bool
	ComputeSize(subroundId int) int
	ResetRoundState()
	IsMultiKeyInConsensusGroup() bool
	IsKeyManagedBySelf(pkBytes []byte) bool
	IncrementRoundsWithoutReceivedMessages(pkBytes []byte)
	GetKeysHandler() consensus.KeysHandler
	Leader() string
}

// RoundStatusHandler encapsulates the methods needed for the status of a subround
type RoundStatusHandler interface {
	Status(subroundId int) SubroundStatus
	SetStatus(subroundId int, subroundStatus SubroundStatus)
	ResetRoundStatus()
}

// RoundThresholdHandler encapsulates the methods needed for the round consensus threshold
type RoundThresholdHandler interface {
	Threshold(subroundId int) int
	SetThreshold(subroundId int, threshold int)
	FallbackThreshold(subroundId int) int
	SetFallbackThreshold(subroundId int, threshold int)
}

// InvalidSignersCache encapsulates the methods needed for a invalid signers cache
type InvalidSignersCache interface {
	AddInvalidSigners(headerHash []byte, invalidSigners []byte, invalidPublicKeys []string)
	CheckKnownInvalidSigners(headerHash []byte, invalidSigners []byte) bool
	Reset()
	IsInterfaceNil() bool
}

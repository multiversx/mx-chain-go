package spos

import (
	"github.com/ElrondNetwork/elastic-indexer-go/workItems"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/redundancy"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ConsensusCoreHandler encapsulates all needed data for the Consensus
type ConsensusCoreHandler interface {
	// Blockchain gets the ChainHandler stored in the ConsensusCore
	Blockchain() data.ChainHandler
	// BlockProcessor gets the BlockProcessor stored in the ConsensusCore
	BlockProcessor() process.BlockProcessor
	// BootStrapper gets the Bootstrapper stored in the ConsensusCore
	BootStrapper() process.Bootstrapper
	// BroadcastMessenger gets the BroadcastMessenger stored in ConsensusCore
	BroadcastMessenger() consensus.BroadcastMessenger
	// Chronology gets the ChronologyHandler stored in the ConsensusCore
	Chronology() consensus.ChronologyHandler
	// GetAntiFloodHandler returns the antiflood handler which will be used in subrounds
	GetAntiFloodHandler() consensus.P2PAntifloodHandler
	// Hasher gets the Hasher stored in the ConsensusCore
	Hasher() hashing.Hasher
	// Marshalizer gets the Marshalizer stored in the ConsensusCore
	Marshalizer() marshal.Marshalizer
	// MultiSigner gets the MultiSigner stored in the ConsensusCore
	MultiSigner() crypto.MultiSigner
	// Rounder gets the Rounder stored in the ConsensusCore
	Rounder() consensus.Rounder
	// ShardCoordinator gets the ShardCoordinator stored in the ConsensusCore
	ShardCoordinator() sharding.Coordinator
	// SyncTimer gets the SyncTimer stored in the ConsensusCore
	SyncTimer() ntp.SyncTimer
	// NodesCoordinator gets the NodesCoordinator stored in the ConsensusCore
	NodesCoordinator() sharding.NodesCoordinator
	// EpochStartRegistrationHandler gets the RegistrationHandler stored in the ConsensusCore
	EpochStartRegistrationHandler() epochStart.RegistrationHandler
	// PrivateKey returns the private key stored in the ConsensusStore used for randomness and leader's signature generation
	PrivateKey() crypto.PrivateKey
	// SingleSigner returns the single signer stored in the ConsensusStore used for randomness and leader's signature generation
	SingleSigner() crypto.SingleSigner
	// PeerHonestyHandler returns the peer honesty handler which will be used in subrounds
	PeerHonestyHandler() consensus.PeerHonestyHandler
	// HeaderSigVerifier returns the sig verifier handler which will be used in subrounds
	HeaderSigVerifier() consensus.HeaderSigVerifier
	// FallbackHeaderValidator returns the fallback header validator handler which will be used in subrounds
	FallbackHeaderValidator() consensus.FallbackHeaderValidator
	// NodeRedundancyHandler returns the node redundancy handler which will be used in subrounds
	NodeRedundancyHandler() redundancy.NodeRedundancyHandler
	// IsInterfaceNil returns true if there is no value under the interface
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
	// IsSubroundSignature returns if the current subround is about signature
	IsSubroundSignature(int) bool
	// IsSubroundStartRound returns if the current subround is about start round
	IsSubroundStartRound(int) bool
	// GetMaxMessagesInARoundPerPeer returns the maximum number of messages a peer can send per round
	GetMaxMessagesInARoundPerPeer() uint32
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// SubroundsFactory encapsulates the methods specifically for a subrounds factory type (bls, bn)
// for different consensus types
type SubroundsFactory interface {
	GenerateSubrounds() error
	IsInterfaceNil() bool
}

// WorkerHandler represents the interface for the SposWorker
type WorkerHandler interface {
	Close() error
	StartWorking()
	// AddReceivedMessageCall adds a new handler function for a received message type
	AddReceivedMessageCall(messageType consensus.MessageType, receivedMessageCall func(cnsDta *consensus.Message) bool)
	// AddReceivedHeaderHandler adds a new handler function for a received header
	AddReceivedHeaderHandler(handler func(data.HeaderHandler))
	// RemoveAllReceivedMessagesCalls removes all the functions handlers
	RemoveAllReceivedMessagesCalls()
	// ProcessReceivedMessage method redirects the received message to the channel which should handle it
	ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
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
	// SetAppStatusHandler sets the status handler object used to collect useful metrics about consensus state machine
	SetAppStatusHandler(ash core.AppStatusHandler) error
	// ResetConsensusMessages resets at the start of each round all the previous consensus messages received
	ResetConsensusMessages()
	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}

// PoolAdder adds data in a key-value pool
type PoolAdder interface {
	Put(key []byte, value interface{}, sizeInBytes int) (evicted bool)
	IsInterfaceNil() bool
}

// RandSeedVerifier encapsulates methods that are check if header rand seed is correct
type RandSeedVerifier interface {
	VerifyRandSeed(header data.HeaderHandler) error
	IsInterfaceNil() bool
}

// HeaderIntegrityVerifier encapsulates methods useful to check that a header's integrity is correct
type HeaderIntegrityVerifier interface {
	Verify(header data.HeaderHandler) error
	IsInterfaceNil() bool
}

// ConsensusDataIndexer defines the actions that a consensus data indexer has to do
type ConsensusDataIndexer interface {
	SaveRoundsInfo(roundsInfos []workItems.RoundInfo)
	IsInterfaceNil() bool
}

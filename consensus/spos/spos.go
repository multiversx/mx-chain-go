package spos

import (
	"encoding/base64"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// ConsensusCoreHandler encapsulates all needed Data for the Consensus
type ConsensusCoreHandler interface {
	// Blockchain gets the ChainHandler stored in the ConsensusCore
	Blockchain() data.ChainHandler
	// BlockProcessor gets the BlockProcessor stored in the ConsensusCore
	BlockProcessor() process.BlockProcessor
	// BootStrapper gets the Bootstrapper stored in the ConsensusCore
	BootStrapper() process.Bootstrapper
	// Chronology gets the ChronologyHandler stored in the ConsensusCore
	Chronology() consensus.ChronologyHandler
	// Hasher gets the Hasher stored in the ConsensusCore
	Hasher() hashing.Hasher
	// Marshalizer gets the Marshalizer stored in the ConsensusCore
	Marshalizer() marshal.Marshalizer
	// MultiSigner gets the MultiSigner stored in the ConsensusCore
	MultiSigner() crypto.MultiSigner
	// Rounder gets the Rounder stored in the ConsensusCore
	Rounder() consensus.Rounder
	// ShardCoordinator gets the Coordinator stored in the ConsensusCore
	ShardCoordinator() sharding.Coordinator
	// SyncTimer gets the SyncTimer stored in the ConsensusCore
	SyncTimer() ntp.SyncTimer
	// ValidatorGroupSelector gets the ValidatorGroupSelector stored in the ConsensusCore
	ValidatorGroupSelector() consensus.ValidatorGroupSelector
	// BlsPrivateKey returns the BLS private key stored in the ConsensusStore
	BlsPrivateKey() crypto.PrivateKey
	// BlsSingleSigner returns the bls single signer stored in the ConsensusStore
	BlsSingleSigner() crypto.SingleSigner
}

//ConsensusService encapsulates the methods specifically for a consensus type (bls, bn)
//and will be used in the sposWorker
type ConsensusService interface {
	//InitReceivedMessages initializes the MessagesType map for all messages for the current ConsensusService
	InitReceivedMessages() map[consensus.MessageType][]*consensus.Message
	//GetStringValue gets the name of the messageType
	GetStringValue(consensus.MessageType) string
	//GetSubroundName gets the subround name for the subround id provided
	GetSubroundName(int) string
	//GetMessageRange provides the MessageType range used in checks by the consensus
	GetMessageRange() []consensus.MessageType
	//CanProceed returns if the current messageType can proceed further if previous subrounds finished
	CanProceed(*ConsensusState, consensus.MessageType) bool
}

//WorkerHandler represents the interface for the SposWorker
type WorkerHandler interface {
	//AddReceivedMessageCall adds a new handler function for a received messege type
	AddReceivedMessageCall(messageType consensus.MessageType, receivedMessageCall func(cnsDta *consensus.Message) bool)
	//RemoveAllReceivedMessagesCalls removes all the functions handlers
	RemoveAllReceivedMessagesCalls()
	//ProcessReceivedMessage method redirects the received message to the channel which should handle it
	ProcessReceivedMessage(message p2p.MessageP2P) error
	//SendConsensusMessage sends the consensus message
	SendConsensusMessage(cnsDta *consensus.Message) bool
	//Extend does an extension for the subround with subroundId
	Extend(subroundId int)
	//GetConsensusStateChangedChannel gets the channel for the consensusStateChanged
	GetConsensusStateChangedChannel() chan bool
	//BroadcastBlock ge
	BroadcastBlock(body data.BodyHandler, header data.HeaderHandler) error
}

// GenerateRandomness is used to generate a random string based on randSeed concatenated with round index
func GenerateRandomness(roundIndex int32, randSeed []byte) string {
	return fmt.Sprintf("%d-%s", roundIndex, toB64(randSeed))
}

// toB64 convert a byte array to a base64 string
func toB64(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}

	return base64.StdEncoding.EncodeToString(buff)
}

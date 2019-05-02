package spos

import (
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

//ConsensusCoreHandler encapsulates all needed Data for the Consensus
type ConsensusCoreHandler interface {
	//Blockchain gets the ChainHandler stored in the ConsensusCore
	Blockchain() data.ChainHandler
	//BlockProcessor gets the BlockProcessor stored in the ConsensusCore
	BlockProcessor() process.BlockProcessor
	//BootStrapper gets the Bootstrapper stored in the ConsensusCore
	BootStrapper() process.Bootstrapper
	//Chronology gets the ChronologyHandler stored in the ConsensusCore
	Chronology() consensus.ChronologyHandler
	//Hasher gets the Hasher stored in the ConsensusCore
	Hasher() hashing.Hasher
	//Marshalizer gets the Marshalizer stored in the ConsensusCore
	Marshalizer() marshal.Marshalizer
	//MultiSigner gets the MultiSigner stored in the ConsensusCore
	MultiSigner() crypto.MultiSigner
	//Rounder gets the Rounder stored in the ConsensusCore
	Rounder() consensus.Rounder
	//ShardCoordinator gets the Coordinator stored in the ConsensusCore
	ShardCoordinator() sharding.Coordinator
	//SyncTimer gets the SyncTimer stored in the ConsensusCore
	SyncTimer() ntp.SyncTimer
	//ValidatorGroupSelector gets the ValidatorGroupSelector stored in the ConsensusCore
	ValidatorGroupSelector() consensus.ValidatorGroupSelector
}

type ConsensusService interface {
	InitReceivedMessages() map[consensus.MessageType][]*consensus.Message
	GetStringValue(consensus.MessageType) string
	GetSubroundName(int) string
	GetMessageRange() []consensus.MessageType
	IsFinished(*ConsensusState, consensus.MessageType) bool
}
type IWorker interface {
	AddReceivedMessageCall(messageType consensus.MessageType, receivedMessageCall func(cnsDta *consensus.Message) bool)
	RemoveAllReceivedMessagesCalls()
	ProcessReceivedMessage(message p2p.MessageP2P) error
	SendConsensusMessage(cnsDta *consensus.Message) bool
	Extend(subroundId int)
	GetConsensusStateChangedChannels() chan bool
	GetBroadcastBlock(body data.BodyHandler, header data.HeaderHandler) error
}

package process

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// TransactionProcessor is the main interface for transaction execution engine
type TransactionProcessor interface {
	SCHandler() func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error
	SetSCHandler(func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error)

	ProcessTransaction(transaction *transaction.Transaction, round uint32) error
}

// BlockProcessor is the main interface for block execution engine
type BlockProcessor interface {
	ProcessBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlock(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountState()
	CreateBlockBody(round uint32, haveTime func() bool) (data.BodyHandler, error)
	RestoreBlockIntoPools(header data.HeaderHandler, body data.BodyHandler) error
	CreateBlockHeader(body data.BodyHandler, round uint32, haveTime func() bool) (data.HeaderHandler, error)
	MarshalizedDataToBroadcast(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[uint32][][]byte, error)
	DecodeBlockBody(dta []byte) data.BodyHandler
	DecodeBlockHeader(dta []byte) data.HeaderHandler
}

// Checker provides functionality to checks the integrity and validity of a data structure
type Checker interface {
	// IntegrityAndValidity does both validity and integrity checks on the data structure
	IntegrityAndValidity(coordinator sharding.Coordinator) error
	// Integrity checks only the integrity of the data
	Integrity(coordinator sharding.Coordinator) error
}

// SigVerifier provides functionality to verify a signature of a signed data structure that holds also the verifying parameters
type SigVerifier interface {
	VerifySig() error
}

// SignedDataValidator provides functionality to check the validity and signature of a data structure
type SignedDataValidator interface {
	SigVerifier
	Checker
}

// HashAccesser interface provides functionality over hashable objects
type HashAccesser interface {
	SetHash([]byte)
	Hash() []byte
}

// InterceptedBlockBody interface provides functionality over intercepted blocks
type InterceptedBlockBody interface {
	Checker
	HashAccesser
	GetUnderlyingObject() interface{}
}

// Bootstrapper is an interface that defines the behaviour of a struct that is able
// to synchronize the node
type Bootstrapper interface {
	AddSyncStateListener(func(bool))
	ShouldSync() bool
	StopSync()
	StartSync()
}

// ForkDetector is an interface that defines the behaviour of a struct that is able
// to detect forks
type ForkDetector interface {
	AddHeader(header data.HeaderHandler, hash []byte, state BlockHeaderState) error
	RemoveHeaders(nonce uint64, hash []byte)
	CheckFork() (bool, uint64)
	GetHighestFinalBlockNonce() uint64
	ProbableHighestNonce() uint64
}

// InterceptorsContainer defines an interceptors holder data type with basic functionality
type InterceptorsContainer interface {
	Get(key string) (Interceptor, error)
	Add(key string, val Interceptor) error
	AddMultiple(keys []string, interceptors []Interceptor) error
	Replace(key string, val Interceptor) error
	Remove(key string)
	Len() int
}

// InterceptorsContainerFactory defines the functionality to create an interceptors container
type InterceptorsContainerFactory interface {
	Create() (InterceptorsContainer, error)
}

// Interceptor defines what a data interceptor should do
// It should also adhere to the p2p.MessageProcessor interface so it can wire to a p2p.Messenger
type Interceptor interface {
	ProcessReceivedMessage(message p2p.MessageP2P) error
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

// ChronologyValidator defines the functionality needed to validate a received header block (shard or metachain)
// from chronology point of view
type ChronologyValidator interface {
	ValidateReceivedBlock(shardID uint32, epoch uint32, nonce uint64, round uint32) error
}

// DataPacker can split a large slice of byte slices in smaller packets
type DataPacker interface {
	PackDataInChunks(data [][]byte, limit int) ([][]byte, error)
}

// BlocksTracker defines the functionality to track all the notarised blocks
type BlocksTracker interface {
	UnnotarisedBlocks() []data.HeaderHandler
	RemoveNotarisedBlocks(headerHandler data.HeaderHandler) error
	AddBlock(headerHandler data.HeaderHandler)
	SetBlockBroadcastRound(nonce uint64, round int32)
	BlockBroadcastRound(nonce uint64) int32
}

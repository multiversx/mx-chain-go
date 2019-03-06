package process

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// TransactionProcessor is the main interface for transaction execution engine
type TransactionProcessor interface {
	SCHandler() func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error
	SetSCHandler(func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error)

	ProcessTransaction(transaction *transaction.Transaction, round int32) error
	SetBalancesToTrie(accBalance map[string]*big.Int) (rootHash []byte, err error)
}

// BlockProcessor is the main interface for block execution engine
type BlockProcessor interface {
	ProcessBlock(blockChain *blockchain.BlockChain, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error
	CommitBlock(blockChain *blockchain.BlockChain, header data.HeaderHandler, body data.BodyHandler) error
	RevertAccountState()
	CreateGenesisBlock(balances map[string]*big.Int) ([]byte, error)
	CreateBlockBody(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (data.BodyHandler, error)
	RemoveBlockInfoFromPool(body data.BodyHandler) error
	GetRootHash() []byte
	CheckBlockValidity(blockChain *blockchain.BlockChain, header data.HeaderHandler, body data.BodyHandler) bool
}

// Checker provides functionality to checks the integrity and validity of a data structure
type Checker interface {
	// IntegrityAndValidity does both validity and integrity checks on the data structure
	IntegrityAndValidity(coordinator sharding.ShardCoordinator) error
	// Integrity checks only the integrity of the data
	Integrity(coordinator sharding.ShardCoordinator) error
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

// IntRandomizer interface provides functionality over generating integer numbers
type IntRandomizer interface {
	Intn(n int) int
}

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
	GetMiniBlocks(hashes [][]byte) block.MiniBlockSlice
}

// TopicResolverSender defines what sending operations are allowed for a topic resolver
type TopicResolverSender interface {
	SendOnRequestTopic(rd *RequestData) error
	Send(buff []byte, peer p2p.PeerID) error
	RequestTopicSuffix() string
}

// Bootstrapper is an interface that defines the behaviour of a struct that is able
// to synchronize the node
type Bootstrapper interface {
	CreateAndCommitEmptyBlock(uint32) (block.Body, *block.Header, error)
	AddSyncStateListener(func(bool))
	ShouldSync() bool
}

// ForkDetector is an interface that defines the behaviour of a struct that is able
// to detect forks
type ForkDetector interface {
	AddHeader(header *block.Header, hash []byte, isReceived bool) error
	RemoveHeaders(nonce uint64)
	CheckFork() bool
}

// Container defines a holder data type with basic functionality
type Container interface {
	Get(key string) (interface{}, error)
	Add(key string, val interface{}) error
	Replace(key string, val interface{}) error
	Remove(key string)
	Len() int
}

// ResolversContainer defines a resolvers holder data type with basic functionality
type ResolversContainer interface {
	Get(key string) (Resolver, error)
	Add(key string, val Resolver) error
	Replace(key string, val Resolver) error
	Remove(key string)
	Len() int
}

// InterceptorsResolversFactory is an interface that defines the behaviour for a factory that
//  can create the needed interceptors and resolvers for the application
type InterceptorsResolversFactory interface {
	CreateInterceptors() error
	CreateResolvers() error
	InterceptorContainer() Container
	ResolverContainer() ResolversContainer
}

type WireMessageHandler interface {
	ConnectedPeers() []p2p.PeerID
	SendToConnectedPeer(topic string, buff []byte, peerID p2p.PeerID) error
}

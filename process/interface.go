package process

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
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
	ProcessBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error
	ProcessAndCommit(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error
	CommitBlock(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error
	RevertAccountState()
	CreateGenesisBlockBody(balances map[string]*big.Int, shardId uint32) (*block.StateBlockBody, error)
	CreateTxBlockBody(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (*block.TxBlockBody, error)
	CreateEmptyBlockBody(shardId uint32, round int32) *block.TxBlockBody
	RemoveBlockTxsFromPool(body *block.TxBlockBody) error
	GetRootHash() []byte
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

// Interceptor defines what a data interceptor should do
type Interceptor interface {
	Name() string
	SetReceivedMessageHandler(handler p2p.TopicValidatorHandler)
	ReceivedMessageHandler() p2p.TopicValidatorHandler
	Marshalizer() marshal.Marshalizer
}

// Resolver is an interface that defines the behaviour of a struct that is able
// to send data requests to other entities and to resolve requests that came from those other entities
type Resolver interface {
	RequestData(rd RequestData) error
	SetResolverHandler(func(rd RequestData) ([]byte, error))
	ResolverHandler() func(rd RequestData) ([]byte, error)
}

// Bootstrapper is an interface that defines the behaviour of a struct that is able
// to synchronize the node
type Bootstrapper interface {
	ShouldSync() bool
}

// ForkDetector is an interface that defines the behaviour of a struct that is able
// to detect forks
type ForkDetector interface {
	AddHeader(header *block.Header, hash []byte, isReceived bool) error
	RemoveHeaders(nonce uint64)
	CheckFork() bool
}

// InterceptorContainer is an interface that defines the beahaviour for a container
//  holding a list of interceptors organized by type
type InterceptorContainer interface {
	Get(key string) (Interceptor, error)
	Add(key string, interceptor Interceptor) error
	Replace(key string, interceptor Interceptor) error
	Remove(key string)
	Len() int
}

// ResolverContainer is an interface that defines the beahaviour for a container
//  holding a list of resolvers organized by type
type ResolverContainer interface {
	Get(key string) (Resolver, error)
	Add(key string, resolver Resolver) error
	Replace(key string, interceptor Resolver) error
	Remove(key string)
	Len() int
}

// ProcessorFactory is an interface that defines the behaviour for a factory that
//  can create the needed interceptors and resolvers for the application
type ProcessorFactory interface {
	CreateInterceptors() error
	CreateResolvers() error
	InterceptorContainer() InterceptorContainer
	ResolverContainer() ResolverContainer
}

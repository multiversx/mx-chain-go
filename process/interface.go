package process

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
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

// blockProcessor is the main interface for block execution engine
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
	CheckBlockValidity(blockChain *blockchain.BlockChain, header *block.Header) bool
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

// TransactionInterceptorAdapter is the interface used in interception of transactions
type TransactionInterceptorAdapter interface {
	Checker
	SigVerifier
	HashAccesser
	p2p.Creator
	RcvShard() uint32
	SndShard() uint32
	IsAddressedToOtherShards() bool
	SetAddressConverter(converter state.AddressConverter)
	AddressConverter() state.AddressConverter
	GetTransaction() *transaction.Transaction
	SingleSignKeyGen() crypto.KeyGenerator
	SetSingleSignKeyGen(generator crypto.KeyGenerator)
	SetTxBuffWithoutSig(txBuffWithoutSig []byte)
	TxBuffWithoutSig() []byte
}

// BlockBodyInterceptorAdapter defines what a block body object should do
type BlockBodyInterceptorAdapter interface {
	Checker
	HashAccesser
	p2p.Creator
	Shard() uint32
	GetUnderlyingObject() interface{}
}

// HeaderInterceptorAdapter is the interface used in interception of headers
type HeaderInterceptorAdapter interface {
	BlockBodyInterceptorAdapter
	SigVerifier
	GetHeader() *block.Header
}

// Interceptor defines what a data interceptor should do
type Interceptor interface {
	Name() string
	SetCheckReceivedObjectHandler(func(newer p2p.Creator, rawData []byte) error)
	CheckReceivedObjectHandler() func(newer p2p.Creator, rawData []byte) error
	Marshalizer() marshal.Marshalizer
}

// Resolver is an interface that defines the behaviour of a struct that is able
// to send data requests to other entities and to resolve requests that came from those other entities
type Resolver interface {
	RequestData(rd RequestData) error
	SetResolverHandler(func(rd RequestData) ([]byte, error))
	ResolverHandler() func(rd RequestData) ([]byte, error)
}

// Bootstraper is an interface that defines the behaviour of a struct that is able
// to syncronize the node
type Bootstraper interface {
	CreateAndCommitEmptyBlock(uint32) (*block.TxBlockBody, *block.Header)
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

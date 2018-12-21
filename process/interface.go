package process

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
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

	ProcessTransaction(transaction *transaction.Transaction) error
	SetBalancesToTrie(accBalance map[string]big.Int) (rootHash []byte, err error)
}

// BlockProcessor is the main interface for block execution engine
type BlockProcessor interface {
	ProcessBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error
	CreateGenesisBlockBody(balances map[string]big.Int, shardId uint32) *block.StateBlockBody
	CreateTxBlockBody(shardId uint32, maxTxInBlock int, haveTime func() bool) (*block.TxBlockBody, error)
	RemoveBlockTxsFromPool(body *block.TxBlockBody) error
	GetRootHash() []byte
	NoShards() uint32
	SetNoShards(uint32)
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

// Hashed interface provides functionality over hashable objects
type Hashed interface {
	SetHash([]byte)
	Hash() []byte
}

// TransactionInterceptorAdapter is the interface used in interception of transactions
type TransactionInterceptorAdapter interface {
	Checker
	SigVerifier
	Hashed
	p2p.Newer
	RcvShard() uint32
	SndShard() uint32
	IsAddressedToOtherShards() bool
	SetAddressConverter(converter state.AddressConverter)
	AddressConverter() state.AddressConverter
	GetTransaction() *transaction.Transaction
	SingleSignKeyGen() crypto.KeyGenerator
	SetSingleSignKeyGen(generator crypto.KeyGenerator)
}

// BlockBodyInterceptorAdapter defines what a block body object should do
type BlockBodyInterceptorAdapter interface {
	Checker
	Hashed
	p2p.Newer
	Shard() uint32
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
	SetCheckReceivedObjectHandler(func(newer p2p.Newer, rawData []byte) bool)
	CheckReceivedObjectHandler() func(newer p2p.Newer, rawData []byte) bool
}

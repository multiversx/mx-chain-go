package process

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
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
	//Check does both validity and integrity checks on the data structure
	Check(processor BlockProcessor) bool
	// Integrity checks only the integrity of the data
	Integrity(processor BlockProcessor) bool
}

// SigVerifier provides functionality to verify a signature of a signed data structure that holds also the verifying parameters
type SigVerifier interface {
	VerifySig() bool
}

// SignedDataValidator provides functionality to check the validity and signature of a data structure
type SignedDataValidator interface {
	SigVerifier
	Checker
}

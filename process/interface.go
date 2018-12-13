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
	CreateGenesisBlock(balances map[string]big.Int, shardId uint32, genTime uint64) *block.StateBlockBody
	CreateTxBlock(nbShards int, shardId uint32, maxTxInBlock int, haveTime func() bool) (*block.TxBlockBody, error)
	RemoveBlockTxsFromPool(body *block.TxBlockBody)
}

// Checker checks the integrity of a data structure
type Checker interface {
	Check() bool
}

// SigVerifier provides functionality to verify a signature
type SigVerifier interface {
	VerifySig() bool
}

// SignedDataValidator provides functionality and check the validity of a block header
type SignedDataValidator interface {
	SigVerifier
	Checker
}

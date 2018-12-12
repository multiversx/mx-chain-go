package execution

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"math/big"
)

// TransactionExecutor is the main interface for transaction execution engine
type TransactionExecutor interface {
	SChandler() func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error
	SetSChandler(func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error)

	RegisterHandler() func(data []byte) error
	SetRegisterHandler(func(data []byte) error)

	UnregisterHandler() func(data []byte) error
	SetUnregisterHandler(func(data []byte) error)

	ProcessTransaction(transaction *transaction.Transaction) error
}

// BlockExecutor is the main interface for block execution engine
type BlockExecutor interface {
	ProcessBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.Block) error
}

// ValidatorSyncer is the main interface for validator syncing engine
type ValidatorSyncer interface {
	AddValidator(nodeId string, stake big.Int)
	RemoveValidator(nodeId string)
}

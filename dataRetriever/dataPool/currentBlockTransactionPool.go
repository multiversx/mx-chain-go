package dataPool

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/dataRetriever"
)

var _ dataRetriever.TransactionCacher = (*transactionMapCacher)(nil)

type transactionMapCacher struct {
	mutTxs      sync.RWMutex
	txsForBlock map[string]data.TransactionHandler
}

// NewCurrentBlockTransactionsPool returns a new transactions pool to be used for the current block
func NewCurrentBlockTransactionsPool() *transactionMapCacher {
	return &transactionMapCacher{
		mutTxs:      sync.RWMutex{},
		txsForBlock: make(map[string]data.TransactionHandler),
	}
}

// Clean creates a new transaction pool
func (tmc *transactionMapCacher) Clean() {
	tmc.mutTxs.Lock()
	tmc.txsForBlock = make(map[string]data.TransactionHandler)
	tmc.mutTxs.Unlock()
}

// GetTx gets the transaction for the given hash
func (tmc *transactionMapCacher) GetTx(txHash []byte) (data.TransactionHandler, error) {
	tmc.mutTxs.RLock()
	defer tmc.mutTxs.RUnlock()

	tx, ok := tmc.txsForBlock[string(txHash)]
	if !ok {
		return nil, dataRetriever.ErrTxNotFoundInBlockPool
	}

	return tx, nil
}

// AddTx adds the transaction for the given hash
func (tmc *transactionMapCacher) AddTx(txHash []byte, tx data.TransactionHandler) {
	if check.IfNil(tx) {
		return
	}

	tmc.mutTxs.Lock()
	tmc.txsForBlock[string(txHash)] = tx
	tmc.mutTxs.Unlock()
}

// IsInterfaceNil returns true if underlying object is nil
func (tmc *transactionMapCacher) IsInterfaceNil() bool {
	return tmc == nil
}

package dataPool

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

var _ dataRetriever.TransactionCacher = (*transactionMapCacher)(nil)

type transactionMapCacher struct {
	mutTxs      sync.RWMutex
	txsForBlock map[string]data.TransactionHandler
}

// NewCurrentBlockPool returns a new pool to be used for current block
func NewCurrentBlockPool() (*transactionMapCacher, error) {
	tmc := &transactionMapCacher{
		mutTxs:      sync.RWMutex{},
		txsForBlock: make(map[string]data.TransactionHandler),
	}

	return tmc, nil
}

// Clean creates a new pool
func (tmc *transactionMapCacher) Clean() {
	tmc.mutTxs.Lock()
	tmc.txsForBlock = make(map[string]data.TransactionHandler)
	tmc.mutTxs.Unlock()
}

// GetTx returns the element saved for the hash
func (tmc *transactionMapCacher) GetTx(txHash []byte) (data.TransactionHandler, error) {
	tmc.mutTxs.RLock()
	defer tmc.mutTxs.RUnlock()

	tx, ok := tmc.txsForBlock[string(txHash)]
	if !ok {
		return nil, dataRetriever.ErrTxNotFoundInBlockPool
	}

	return tx, nil
}

// AddTx writes the tx to the map
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

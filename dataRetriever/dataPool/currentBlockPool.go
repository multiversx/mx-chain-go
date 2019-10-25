package dataPool

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

type currentBlockPool struct {
	mutTxs      sync.RWMutex
	txsForBlock map[string]data.TransactionHandler
}

// NewCurrentBlockPool returns a new pool to be used for current block
func NewCurrentBlockPool() (*currentBlockPool, error) {
	currPool := &currentBlockPool{
		mutTxs:      sync.RWMutex{},
		txsForBlock: make(map[string]data.TransactionHandler),
	}

	return currPool, nil
}

// Clean creates a new pool
func (c *currentBlockPool) Clean() {
	c.mutTxs.Lock()
	c.txsForBlock = make(map[string]data.TransactionHandler)
	c.mutTxs.Unlock()
}

// GetTx returns the element saved for the hash
func (c *currentBlockPool) GetTx(txHash []byte) (data.TransactionHandler, error) {
	c.mutTxs.RLock()
	defer c.mutTxs.RUnlock()

	tx, ok := c.txsForBlock[string(txHash)]
	if !ok {
		return nil, dataRetriever.ErrNilValue
	}
	return tx, nil
}

// AddTx writes the tx to the map
func (c *currentBlockPool) AddTx(txHash []byte, tx data.TransactionHandler) {
	c.mutTxs.Lock()
	c.txsForBlock[string(txHash)] = tx
	c.mutTxs.Unlock()
}

// IsInterfaceNil returns true if underlying object is nil
func (c *currentBlockPool) IsInterfaceNil() bool {
	if c == nil {
		return true
	}
	return false
}

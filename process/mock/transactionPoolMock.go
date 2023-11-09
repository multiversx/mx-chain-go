package mock

import (
	"github.com/multiversx/mx-chain-go/storage"
)

// TransactionPoolMock -
type TransactionPoolMock struct {
	RegisterTransactionHandlerCalled func(transactionHandler func(txHash []byte))
	RemoveTransactionsFromPoolCalled func(txHashes [][]byte, destShardID uint32)
	MiniPoolTxStoreCalled            func(shardID uint32) (c storage.Cacher)
}

// RegisterTransactionHandler -
func (tpm *TransactionPoolMock) RegisterTransactionHandler(transactionHandler func(txHash []byte)) {
	tpm.RegisterTransactionHandlerCalled(transactionHandler)
}

// RemoveTransactionsFromPool -
func (tpm *TransactionPoolMock) RemoveTransactionsFromPool(txHashes [][]byte, destShardID uint32) {
	tpm.RemoveTransactionsFromPoolCalled(txHashes, destShardID)
}

// MiniPoolTxStore -
func (tpm *TransactionPoolMock) MiniPoolTxStore(shardID uint32) (c storage.Cacher) {
	return tpm.MiniPoolTxStoreCalled(shardID)
}

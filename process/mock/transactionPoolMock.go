package mock

import (
	"github.com/ElrondNetwork/elrond-go/storage"
)

type TransactionPoolMock struct {
	RegisterTransactionHandlerCalled func(transactionHandler func(txHash []byte))
	RemoveTransactionsFromPoolCalled func(txHashes [][]byte, destShardID uint32)
	MiniPoolTxStoreCalled            func(shardID uint32) (c storage.Cacher)
}

func (tpm *TransactionPoolMock) RegisterTransactionHandler(transactionHandler func(txHash []byte)) {
	tpm.RegisterTransactionHandlerCalled(transactionHandler)
}

func (tpm *TransactionPoolMock) RemoveTransactionsFromPool(txHashes [][]byte, destShardID uint32) {
	tpm.RemoveTransactionsFromPoolCalled(txHashes, destShardID)
}

func (tpm *TransactionPoolMock) MiniPoolTxStore(shardID uint32) (c storage.Cacher) {
	return tpm.MiniPoolTxStoreCalled(shardID)
}

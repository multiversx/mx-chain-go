package txcache

import (
	"github.com/ElrondNetwork/elrond-go/storage"
)

type txCache interface {
	storage.Cacher

	AddTx(tx *WrappedTransaction) (ok bool, added bool)
	GetByTxHash(txHash []byte) (*WrappedTransaction, bool)
	RemoveTxByHash(txHash []byte) error
	CountTx() int64
	ForEachTransaction(function ForEachTransaction)
	SelectTransactions(numRequested int, batchSizePerSender int) []*WrappedTransaction
}

type scoreComputer interface {
	computeScore(scoreParams senderScoreParams) uint32
}

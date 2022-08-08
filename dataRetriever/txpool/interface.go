package txpool

import (
	"github.com/ElrondNetwork/elrond-go-storage/txcache"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type txCache interface {
	storage.Cacher

	AddTx(tx *txcache.WrappedTransaction) (ok bool, added bool)
	GetByTxHash(txHash []byte) (*txcache.WrappedTransaction, bool)
	RemoveTxByHash(txHash []byte) bool
	ImmunizeTxsAgainstEviction(keys [][]byte)
	ForEachTransaction(function txcache.ForEachTransaction)
	NumBytes() int
	Diagnose(deep bool)
	GetTransactionsPoolForSender(sender string) []*txcache.WrappedTransaction
}

package txpool

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/txcache"
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

type txGasHandler interface {
	ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int
	IsInterfaceNil() bool
}

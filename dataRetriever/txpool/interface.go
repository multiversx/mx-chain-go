package txpool

import (
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/txcache"
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
	OnProposedBlock(blockHash []byte, blockBody *block.Body, blockHeader data.HeaderHandler, accountsProvider common.AccountNonceAndBalanceProvider, blockchainInfo common.BlockchainInfo) error
	OnExecutedBlock(blockHeader data.HeaderHandler) error
	Cleanup(accountsProvider common.AccountNonceProvider, randomness uint64, maxNum int, cleanupLoopMaximumDuration time.Duration) uint64
}

type txGasHandler interface {
	ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int
	IsInterfaceNil() bool
}

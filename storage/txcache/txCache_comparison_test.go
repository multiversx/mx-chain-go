package txcache

import (
	"container/list"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

var maxSize uint32 = 250000
var neededTransactions = 15000
var iterationsToRun int = 4
var passBatchSize = 2

func Test_FIFOSharded(t *testing.T) {
	watch := stopWatch{}

	for index := 0; index < iterationsToRun; index++ {
		transactions := createUniformlyDistributedTransactions(1000, 1000, true)
		cache := prepareCacheFIFOSharded(16, transactions)
		watch.start()
		orderedTxs, _, _ := preprocess.SortTxByNonce(cache)
		fmt.Println("orderedTxs", len(orderedTxs))
		watch.pause()
	}

	fmt.Println(">>>", "Uniform distribution;", iterationsToRun, "iterations;", "duration:", watch.format())

	watch = stopWatch{}

	for index := 0; index < iterationsToRun; index++ {
		transactions := createUniformlyDistributedTransactions(1000, 1000, false)
		cache := prepareCacheFIFOSharded(16, transactions)
		watch.start()
		orderedTxs, _, _ := preprocess.SortTxByNonce(cache)
		fmt.Println("orderedTxs", len(orderedTxs))
		watch.pause()
	}

	fmt.Println(">>>", "Uniform distribution, REVERSED nonces;", iterationsToRun, "iterations;", "duration:", watch.format())
}

func prepareCacheFIFOSharded(noShards int, transactions *list.List) storage.Cacher {
	cache, _ := storageUnit.NewCache(storageUnit.FIFOShardedCache, maxSize, uint32(noShards))

	for e := transactions.Front(); e != nil; e = e.Next() {
		item := e.Value.(*txWithHash)
		cache.Put(item.Hash, item.Transaction)
	}

	return cache
}

func Test_TxCache(t *testing.T) {
	watch := stopWatch{}

	for index := 0; index < iterationsToRun; index++ {
		transactions := createUniformlyDistributedTransactions(1000, 1000, true)
		cache := prepareTxCache(16, transactions)
		watch.start()
		_ = cache.GetSorted(neededTransactions, passBatchSize)
		watch.pause()
	}

	fmt.Println(">>>", "Uniform distribution;", iterationsToRun, "iterations;", "duration:", watch.format())

	for index := 0; index < iterationsToRun; index++ {
		transactions := createUniformlyDistributedTransactions(1000, 1000, false)
		cache := prepareTxCache(16, transactions)
		watch.start()
		_ = cache.GetSorted(neededTransactions, passBatchSize)
		watch.pause()
	}

	fmt.Println(">>>", "Uniform distribution REVERSED nonces;", iterationsToRun, "iterations;", "duration:", watch.format())
}

func prepareTxCache(noShards int, transactions *list.List) *TxCache {
	cache := NewTxCache(int(maxSize), noShards)

	for e := transactions.Front(); e != nil; e = e.Next() {
		item := e.Value.(*txWithHash)
		cache.AddTx(item.Hash, item.Transaction)
	}

	return cache
}

func createUniformlyDistributedTransactions(noSenders int, noTxPerSender int, goodNoncesOrder bool) *list.List {
	result := list.New()

	for senderTag := 0; senderTag < noSenders; senderTag++ {
		sender := createRandom32()

		if goodNoncesOrder {
			for txNonce := 0; txNonce < noTxPerSender; txNonce++ {
				result.PushBack(createTxWithHash(sender, txNonce))
			}
		} else {
			for txNonce := noTxPerSender; txNonce > 0; txNonce-- {
				result.PushBack(createTxWithHash(sender, txNonce))
			}
		}
	}

	return result
}

type txWithHash struct {
	Transaction *transaction.Transaction
	Hash        []byte
}

func createTxWithHash(sender []byte, nonce int) *txWithHash {
	txHash := createRandom32()

	tx := &txWithHash{
		Transaction: &transaction.Transaction{
			SndAddr: sender,
			Nonce:   uint64(nonce),
		},

		Hash: txHash,
	}

	return tx
}

func createRandom32() []byte {
	bytes := make([]byte, 32)
	_, _ = rand.Reader.Read(bytes)
	return bytes
}

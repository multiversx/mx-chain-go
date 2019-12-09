package txcache

import (
	"encoding/binary"
	"fmt"
	"math"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/stretchr/testify/assert"
)

func Test_AddTx(t *testing.T) {
	cache := NewTxCache(250000, 16)

	txHash := []byte("hash-1")
	tx := createTx("alice", 1)

	cache.AddTx(txHash, tx)
	foundTx, ok := cache.GetByTxHash(txHash)

	assert.True(t, ok)
	assert.Equal(t, tx, foundTx)
}

func Test_RemoveByTxHash(t *testing.T) {
	cache := NewTxCache(250000, 16)

	txHash := []byte("hash-1")
	tx := createTx("alice", 1)

	cache.AddTx(txHash, tx)
	cache.RemoveTxByHash(txHash)
	foundTx, ok := cache.GetByTxHash(txHash)

	assert.False(t, ok)
	assert.Nil(t, foundTx)
}

func Test_GetSorted_Dummy(t *testing.T) {
	cache := NewTxCache(250000, 16)

	cache.AddTx([]byte("hash-alice-4"), createTx("alice", 4))
	cache.AddTx([]byte("hash-alice-3"), createTx("alice", 3))
	cache.AddTx([]byte("hash-alice-2"), createTx("alice", 2))
	cache.AddTx([]byte("hash-alice-1"), createTx("alice", 1))
	cache.AddTx([]byte("hash-bob-7"), createTx("bob", 7))
	cache.AddTx([]byte("hash-bob-6"), createTx("bob", 6))
	cache.AddTx([]byte("hash-bob-5"), createTx("bob", 5))
	cache.AddTx([]byte("hash-carol-1"), createTx("carol", 1))

	sorted := cache.GetSorted(10, 2)
	assert.Len(t, sorted, 8)
}

func Test_GetSorted(t *testing.T) {
	cache := NewTxCache(250000, 16)

	// For "noSenders" senders, add "noTransactions" transactions,
	// in reversed-nonce order.
	// Total of "noSenders" * "noTransactions" transactions in the cache.
	noSenders := 1000
	noTransactionsPerSender := 100
	noTotalTransactions := noSenders * noTransactionsPerSender
	noRequestedTransactions := math.MaxInt16

	for senderTag := 0; senderTag < noSenders; senderTag++ {
		sender := fmt.Sprintf("sender%d", senderTag)

		for txNonce := noTransactionsPerSender; txNonce > 0; txNonce-- {
			txHash := fmt.Sprintf("hash%d%d", senderTag, txNonce)
			tx := createTx(sender, uint64(txNonce))
			cache.AddTx([]byte(txHash), tx)
		}
	}

	assert.Equal(t, int64(noTotalTransactions), cache.CountTx())

	sorted := cache.GetSorted(noRequestedTransactions, 2)

	assert.Len(t, sorted, core.MinInt(noRequestedTransactions, noTotalTransactions))

	// Check order
	nonces := make(map[string]uint64, noSenders)
	for _, tx := range sorted {
		nonce := tx.GetNonce()
		sender := string(tx.GetSndAddress())
		previousNonce := nonces[sender]

		assert.LessOrEqual(t, previousNonce, nonce)
		nonces[sender] = nonce
	}
}

func Test_AddManyTransactionsToCacheWithEviction_UniformDistribution(t *testing.T) {
	logger.SetLogLevel("txcache/eviction:DEBUG")

	cache := NewTxCache(250000, 1)
	config := EvictionStrategyConfig{CountThreshold: 240000, NoOldestSendersToEvict: 10, ManyTransactionsForASender: 1000, PartOfManyTransactionsOfASender: 250}
	cache.EvictionStrategy = NewEvictionStrategy(cache, config)

	noSenders := 5000
	noTransactionsPerSender := 100

	for senderTag := 0; senderTag < noSenders; senderTag++ {
		sender := createFakeSenderAddress(senderTag)

		for txNonce := noTransactionsPerSender; txNonce > 0; txNonce-- {
			txHash := createFakeTxHash(sender, txNonce)
			tx := createTx(string(sender), uint64(txNonce))
			cache.AddTx([]byte(txHash), tx)
		}
	}

	assert.LessOrEqual(t, cache.CountTx(), int64(240000))
}

func createTx(sender string, nonce uint64) *transaction.Transaction {
	return &transaction.Transaction{
		SndAddr: []byte(sender),
		Nonce:   nonce,
	}
}

func createFakeSenderAddress(senderTag int) []byte {
	bytes := make([]byte, 32)
	binary.LittleEndian.PutUint64(bytes, uint64(senderTag))
	binary.LittleEndian.PutUint64(bytes[24:], uint64(senderTag))
	return bytes
}

func createFakeTxHash(fakeSenderAddress []byte, nonce int) []byte {
	bytes := make([]byte, 32)
	copy(bytes, fakeSenderAddress)
	binary.LittleEndian.PutUint64(bytes[8:], uint64(nonce))
	binary.LittleEndian.PutUint64(bytes[16:], uint64(nonce))
	return bytes
}


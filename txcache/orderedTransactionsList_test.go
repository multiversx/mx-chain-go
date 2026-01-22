package txcache

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewOrderedTransactionsList(t *testing.T) {
	otl := newOrderedTransactionsList()
	require.NotNil(t, otl)
	require.Equal(t, 0, otl.len())
	require.Equal(t, 0, len(otl.getAll()))
}

func TestOrderedTransactionsList_Insert(t *testing.T) {
	t.Run("ordered by nonce ascending", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)
		tx2 := createTx([]byte("tx2"), "sender", 2)
		tx3 := createTx([]byte("tx3"), "sender", 3)

		require.True(t, otl.insert(tx1))
		require.True(t, otl.insert(tx2))
		require.True(t, otl.insert(tx3))

		require.Equal(t, 3, otl.len())
		require.Equal(t, []*WrappedTransaction{tx1, tx2, tx3}, otl.getAll())
	})

	t.Run("ordered by nonce descending (insert reverse)", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)
		tx2 := createTx([]byte("tx2"), "sender", 2)
		tx3 := createTx([]byte("tx3"), "sender", 3)

		require.True(t, otl.insert(tx3))
		require.True(t, otl.insert(tx2))
		require.True(t, otl.insert(tx1))

		require.Equal(t, 3, otl.len())
		require.Equal(t, []*WrappedTransaction{tx1, tx2, tx3}, otl.getAll())
	})

	t.Run("mixed nonce and gas price", func(t *testing.T) {
		otl := newOrderedTransactionsList()

		// Expected Order:
		// 1. Nonce 1, Gas 2000
		// 2. Nonce 1, Gas 1000
		// 3. Nonce 2, Gas 500
		// 4. Nonce 3, Gas 3000

		tx1HighGas := createTx([]byte("tx1H"), "sender", 1).withGasPrice(2000)
		tx1LowGas := createTx([]byte("tx1L"), "sender", 1).withGasPrice(1000)
		tx2 := createTx([]byte("tx2"), "sender", 2).withGasPrice(500)
		tx3 := createTx([]byte("tx3"), "sender", 3).withGasPrice(3000)

		// Insert in random order
		require.True(t, otl.insert(tx2))
		require.True(t, otl.insert(tx1LowGas))
		require.True(t, otl.insert(tx3))
		require.True(t, otl.insert(tx1HighGas))

		expected := []*WrappedTransaction{tx1HighGas, tx1LowGas, tx2, tx3}
		require.Equal(t, expected, otl.getAll())
	})

	t.Run("ordered by gas price descending (same nonce)", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 10).withGasPrice(100)
		tx2 := createTx([]byte("tx2"), "sender", 10).withGasPrice(200)
		tx3 := createTx([]byte("tx3"), "sender", 10).withGasPrice(300)

		// Insert low to high (should end up high to low)
		otl.insert(tx1)
		otl.insert(tx2)
		otl.insert(tx3)

		require.Equal(t, []*WrappedTransaction{tx3, tx2, tx1}, otl.getAll())
	})

	t.Run("random insertion mixed", func(t *testing.T) {
		otl := newOrderedTransactionsList()

		tx1 := createTx([]byte("tx1"), "sender", 1)
		tx2 := createTx([]byte("tx2"), "sender", 2)
		tx3 := createTx([]byte("tx3"), "sender", 3)

		require.True(t, otl.insert(tx2))
		require.True(t, otl.insert(tx3))
		require.True(t, otl.insert(tx1))

		require.Equal(t, 3, otl.len())
		require.Equal(t, []*WrappedTransaction{tx1, tx2, tx3}, otl.getAll())
	})

	t.Run("duplicates are rejected", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)

		require.True(t, otl.insert(tx1))
		require.False(t, otl.insert(tx1))

		tx1Copy := createTx([]byte("tx1"), "sender", 1)
		require.False(t, otl.insert(tx1Copy))

		require.Equal(t, 1, otl.len())
	})
}

func TestOrderedTransactionsList_RemoveAt(t *testing.T) {
	otl := newOrderedTransactionsList()
	tx1 := createTx([]byte("tx1"), "sender", 1)
	tx2 := createTx([]byte("tx2"), "sender", 2)
	tx3 := createTx([]byte("tx3"), "sender", 3)

	otl.insert(tx1)
	otl.insert(tx2)
	otl.insert(tx3)

	// Remove middle
	removed := otl.removeAt(1)
	require.Equal(t, tx2, removed)
	require.Equal(t, []*WrappedTransaction{tx1, tx3}, otl.getAll())

	// Remove start
	removed = otl.removeAt(0)
	require.Equal(t, tx1, removed)
	require.Equal(t, []*WrappedTransaction{tx3}, otl.getAll())

	// Remove end
	removed = otl.removeAt(0)
	require.Equal(t, tx3, removed)
	require.Equal(t, 0, otl.len())

	// Out of bounds
	require.Nil(t, otl.removeAt(0))
	require.Nil(t, otl.removeAt(-1))
}

func TestOrderedTransactionsList_RemoveAfterNonce(t *testing.T) {
	t.Run("remove some from end", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)
		tx2 := createTx([]byte("tx2"), "sender", 2)
		tx3 := createTx([]byte("tx3"), "sender", 3)
		tx4 := createTx([]byte("tx4"), "sender", 4)
		otl.insert(tx1)
		otl.insert(tx2)
		otl.insert(tx3)
		otl.insert(tx4)

		removed := otl.removeAfterNonce(3)
		require.Equal(t, []*WrappedTransaction{tx3, tx4}, removed)
		require.Equal(t, []*WrappedTransaction{tx1, tx2}, otl.getAll())
	})

	t.Run("remove all", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)
		otl.insert(tx1)

		removed := otl.removeAfterNonce(0)
		require.Equal(t, []*WrappedTransaction{tx1}, removed)
		require.Equal(t, 0, otl.len())
	})

	t.Run("remove none (too high)", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)
		otl.insert(tx1)

		removed := otl.removeAfterNonce(2)
		require.Nil(t, removed)
		require.Equal(t, 1, otl.len())
	})

	t.Run("remove empty list", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		removed := otl.removeAfterNonce(0)
		require.Nil(t, removed)
	})
}

func TestOrderedTransactionsList_RemoveBeforeNonce(t *testing.T) {
	t.Run("remove some from start", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)
		tx2 := createTx([]byte("tx2"), "sender", 2)
		tx3 := createTx([]byte("tx3"), "sender", 3)
		tx4 := createTx([]byte("tx4"), "sender", 4)
		otl.insert(tx1)
		otl.insert(tx2)
		otl.insert(tx3)
		otl.insert(tx4)

		// Remove <= 2
		removed := otl.removeBeforeNonce(2)
		require.Equal(t, []*WrappedTransaction{tx1, tx2}, removed)
		require.Equal(t, []*WrappedTransaction{tx3, tx4}, otl.getAll())
	})

	t.Run("remove all", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)
		otl.insert(tx1)

		removed := otl.removeBeforeNonce(10)
		require.Equal(t, []*WrappedTransaction{tx1}, removed)
		require.Equal(t, 0, otl.len())
	})

	t.Run("remove none (too low)", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 10)
		otl.insert(tx1)

		removed := otl.removeBeforeNonce(5)
		require.Nil(t, removed)
		require.Equal(t, 1, otl.len())
	})

	t.Run("remove matches exactly", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 10)
		otl.insert(tx1)

		removed := otl.removeBeforeNonce(10)
		require.Equal(t, []*WrappedTransaction{tx1}, removed)
		require.Equal(t, 0, otl.len())
	})
}

func TestOrderedTransactionsList_Get(t *testing.T) {
	otl := newOrderedTransactionsList()
	tx1 := createTx([]byte("tx1"), "sender", 1)
	tx2 := createTx([]byte("tx2"), "sender", 2)
	tx3 := createTx([]byte("tx3"), "sender", 3)
	otl.insert(tx1)
	otl.insert(tx2)
	otl.insert(tx3)

	require.Equal(t, tx1, otl.get(0))
	require.Equal(t, tx2, otl.get(1))
	require.Equal(t, tx3, otl.get(2))

	require.Nil(t, otl.get(3))
	require.Nil(t, otl.get(-1))
}

func TestIsGreater(t *testing.T) {
	txHighPrice := createTx([]byte("high"), "sender", 1).withGasPrice(2000)
	txLowPrice := createTx([]byte("low"), "sender", 1).withGasPrice(1000)

	require.Equal(t, -1, compareTxs(txHighPrice, txLowPrice))
	require.False(t, isGreater(txHighPrice, txLowPrice))
	require.True(t, isGreater(txLowPrice, txHighPrice))
}

func TestOrderedTransactionsList_MemorySafety_Fuzz(t *testing.T) {
	otl := newOrderedTransactionsList()

	// Fuzz-like scenario:
	// Insert 1000 items with random nonces/prices
	// Remove random chunks
	// Verify sorting is maintained

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	numItems := 1000
	for i := 0; i < numItems; i++ {
		nonce := uint64(rnd.Intn(100))
		gas := uint64(rnd.Intn(1000) + 1)
		tx := createTx([]byte(fmt.Sprintf("tx-%d", i)), "sender", nonce).withGasPrice(gas)
		otl.insert(tx)
	}

	// Check integrity
	items := otl.getAll()
	for i := 0; i < len(items)-1; i++ {
		// items[i] should be <= items[i+1]
		// compareTxs(items[i], items[i+1]) should be <= 0
		cmp := compareTxs(items[i], items[i+1])
		require.True(t, cmp <= 0, "List validation failed at index %d: nonce %d vs %d", i, items[i].Tx.GetNonce(), items[i+1].Tx.GetNonce())
	}

	// Remove some
	otl.removeBeforeNonce(50)

	// Check again
	items = otl.getAll()
	for i := 0; i < len(items)-1; i++ {
		cmp := compareTxs(items[i], items[i+1])
		require.True(t, cmp <= 0, "List validation failed after remove at index %d", i)
	}

	// Check that we don't have dangling low nonces
	if len(items) > 0 {
		require.True(t, items[0].Tx.GetNonce() > 50, "Expected all nonces > 50, got %d", items[0].Tx.GetNonce())
	}
}

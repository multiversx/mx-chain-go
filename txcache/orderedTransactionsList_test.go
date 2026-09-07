package txcache

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// insertTx is a test helper that inserts a transaction using findInsertionIndex + insertAt
func (otl *orderedTransactionsList) insertTx(tx *WrappedTransaction) bool {
	index := otl.findInsertionIndex(tx)
	return otl.insertAt(tx, index)
}

func TestNewOrderedTransactionsList(t *testing.T) {
	otl := newOrderedTransactionsList()
	require.NotNil(t, otl)
	require.Equal(t, 0, otl.len())
	require.Equal(t, 0, len(otl.getAll()))
}

func TestOrderedTransactionsList_InsertTx(t *testing.T) {
	t.Run("ordered by nonce ascending", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)
		tx2 := createTx([]byte("tx2"), "sender", 2)
		tx3 := createTx([]byte("tx3"), "sender", 3)

		require.True(t, otl.insertTx(tx1))
		require.True(t, otl.insertTx(tx2))
		require.True(t, otl.insertTx(tx3))

		require.Equal(t, 3, otl.len())
		require.Equal(t, []*WrappedTransaction{tx1, tx2, tx3}, otl.getAll())
	})

	t.Run("ordered by nonce descending (insert reverse)", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)
		tx2 := createTx([]byte("tx2"), "sender", 2)
		tx3 := createTx([]byte("tx3"), "sender", 3)

		require.True(t, otl.insertTx(tx3))
		require.True(t, otl.insertTx(tx2))
		require.True(t, otl.insertTx(tx1))

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
		require.True(t, otl.insertTx(tx2))
		require.True(t, otl.insertTx(tx1LowGas))
		require.True(t, otl.insertTx(tx3))
		require.True(t, otl.insertTx(tx1HighGas))

		expected := []*WrappedTransaction{tx1HighGas, tx1LowGas, tx2, tx3}
		require.Equal(t, expected, otl.getAll())
	})

	t.Run("ordered by gas price descending (same nonce)", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 10).withGasPrice(100)
		tx2 := createTx([]byte("tx2"), "sender", 10).withGasPrice(200)
		tx3 := createTx([]byte("tx3"), "sender", 10).withGasPrice(300)

		// Insert low to high (should end up high to low)
		otl.insertTx(tx1)
		otl.insertTx(tx2)
		otl.insertTx(tx3)

		require.Equal(t, []*WrappedTransaction{tx3, tx2, tx1}, otl.getAll())
	})

	t.Run("random insertion mixed", func(t *testing.T) {
		otl := newOrderedTransactionsList()

		tx1 := createTx([]byte("tx1"), "sender", 1)
		tx2 := createTx([]byte("tx2"), "sender", 2)
		tx3 := createTx([]byte("tx3"), "sender", 3)

		require.True(t, otl.insertTx(tx2))
		require.True(t, otl.insertTx(tx3))
		require.True(t, otl.insertTx(tx1))

		require.Equal(t, 3, otl.len())
		require.Equal(t, []*WrappedTransaction{tx1, tx2, tx3}, otl.getAll())
	})

	t.Run("duplicates are rejected", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)

		require.True(t, otl.insertTx(tx1))
		require.False(t, otl.insertTx(tx1))

		tx1Copy := createTx([]byte("tx1"), "sender", 1)
		require.False(t, otl.insertTx(tx1Copy))

		require.Equal(t, 1, otl.len())
	})
}

func TestOrderedTransactionsList_RemoveAt(t *testing.T) {
	otl := newOrderedTransactionsList()
	tx1 := createTx([]byte("tx1"), "sender", 1)
	tx2 := createTx([]byte("tx2"), "sender", 2)
	tx3 := createTx([]byte("tx3"), "sender", 3)

	otl.insertTx(tx1)
	otl.insertTx(tx2)
	otl.insertTx(tx3)

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
		otl.insertTx(tx1)
		otl.insertTx(tx2)
		otl.insertTx(tx3)
		otl.insertTx(tx4)

		removed := otl.removeAfterNonce(3)
		require.Equal(t, []*WrappedTransaction{tx3, tx4}, removed)
		require.Equal(t, []*WrappedTransaction{tx1, tx2}, otl.getAll())
	})

	t.Run("remove all", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)
		otl.insertTx(tx1)

		removed := otl.removeAfterNonce(0)
		require.Equal(t, []*WrappedTransaction{tx1}, removed)
		require.Equal(t, 0, otl.len())
	})

	t.Run("remove none (too high)", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)
		otl.insertTx(tx1)

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
		otl.insertTx(tx1)
		otl.insertTx(tx2)
		otl.insertTx(tx3)
		otl.insertTx(tx4)

		// Remove <= 2
		removed := otl.removeBeforeNonce(2)
		require.Equal(t, []*WrappedTransaction{tx1, tx2}, removed)
		require.Equal(t, []*WrappedTransaction{tx3, tx4}, otl.getAll())
	})

	t.Run("remove all", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)
		otl.insertTx(tx1)

		removed := otl.removeBeforeNonce(10)
		require.Equal(t, []*WrappedTransaction{tx1}, removed)
		require.Equal(t, 0, otl.len())
	})

	t.Run("remove none (too low)", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 10)
		otl.insertTx(tx1)

		removed := otl.removeBeforeNonce(5)
		require.Nil(t, removed)
		require.Equal(t, 1, otl.len())
	})

	t.Run("remove matches exactly", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 10)
		otl.insertTx(tx1)

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
	otl.insertTx(tx1)
	otl.insertTx(tx2)
	otl.insertTx(tx3)

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
		otl.insertTx(tx)
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

func TestOrderedTransactionsList_GetAllFromIndex(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		result := otl.getAllFromIndex(0)
		require.Equal(t, 0, len(result))
	})

	t.Run("start index 0", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)
		tx2 := createTx([]byte("tx2"), "sender", 2)
		tx3 := createTx([]byte("tx3"), "sender", 3)

		otl.insertTx(tx1)
		otl.insertTx(tx2)
		otl.insertTx(tx3)

		result := otl.getAllFromIndex(0)
		require.Equal(t, 3, len(result))
		require.Equal(t, tx1, result[0])
		require.Equal(t, tx2, result[1])
		require.Equal(t, tx3, result[2])
	})

	t.Run("start index 1", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)
		tx2 := createTx([]byte("tx2"), "sender", 2)
		tx3 := createTx([]byte("tx3"), "sender", 3)

		otl.insertTx(tx1)
		otl.insertTx(tx2)
		otl.insertTx(tx3)

		result := otl.getAllFromIndex(1)
		require.Equal(t, 2, len(result))
		require.Equal(t, tx2, result[0])
		require.Equal(t, tx3, result[1])
	})

	t.Run("start index beyond list", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)
		otl.insertTx(tx1)

		result := otl.getAllFromIndex(5)
		require.Equal(t, 0, len(result))
	})

	t.Run("negative start index", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 1)
		tx2 := createTx([]byte("tx2"), "sender", 2)

		otl.insertTx(tx1)
		otl.insertTx(tx2)

		result := otl.getAllFromIndex(-1)
		require.Equal(t, 2, len(result))
	})
}

func TestOrderedTransactionsList_FindIndexByNonce(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		require.Equal(t, 0, otl.findIndexByNonce(10))
	})

	t.Run("exact match", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		otl.insertTx(createTx([]byte("tx1"), "sender", 10))
		otl.insertTx(createTx([]byte("tx2"), "sender", 20))
		otl.insertTx(createTx([]byte("tx3"), "sender", 30))

		require.Equal(t, 0, otl.findIndexByNonce(10))
		require.Equal(t, 1, otl.findIndexByNonce(20))
		require.Equal(t, 2, otl.findIndexByNonce(30))
	})

	t.Run("nonce between existing", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		otl.insertTx(createTx([]byte("tx1"), "sender", 10))
		otl.insertTx(createTx([]byte("tx2"), "sender", 20))
		otl.insertTx(createTx([]byte("tx3"), "sender", 30))

		// Should return index of first tx with nonce >= 15 (which is tx2 at index 1)
		require.Equal(t, 1, otl.findIndexByNonce(15))
		// Should return index of first tx with nonce >= 25 (which is tx3 at index 2)
		require.Equal(t, 2, otl.findIndexByNonce(25))
	})

	t.Run("nonce below all", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		otl.insertTx(createTx([]byte("tx1"), "sender", 10))
		otl.insertTx(createTx([]byte("tx2"), "sender", 20))

		require.Equal(t, 0, otl.findIndexByNonce(5))
	})

	t.Run("nonce above all", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		otl.insertTx(createTx([]byte("tx1"), "sender", 10))
		otl.insertTx(createTx([]byte("tx2"), "sender", 20))

		require.Equal(t, 2, otl.findIndexByNonce(100))
	})
}

func TestOrderedTransactionsList_FindInsertionIndex(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx := createTx([]byte("tx"), "sender", 10)
		require.Equal(t, 0, otl.findInsertionIndex(tx))
	})

	t.Run("insert at end", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		otl.insertTx(createTx([]byte("tx1"), "sender", 10))
		otl.insertTx(createTx([]byte("tx2"), "sender", 20))

		tx := createTx([]byte("tx3"), "sender", 30)
		require.Equal(t, 2, otl.findInsertionIndex(tx))
	})

	t.Run("insert at beginning", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		otl.insertTx(createTx([]byte("tx1"), "sender", 20))
		otl.insertTx(createTx([]byte("tx2"), "sender", 30))

		tx := createTx([]byte("tx3"), "sender", 10)
		require.Equal(t, 0, otl.findInsertionIndex(tx))
	})

	t.Run("insert in middle", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		otl.insertTx(createTx([]byte("tx1"), "sender", 10))
		otl.insertTx(createTx([]byte("tx2"), "sender", 30))

		tx := createTx([]byte("tx3"), "sender", 20)
		require.Equal(t, 1, otl.findInsertionIndex(tx))
	})
}

func TestOrderedTransactionsList_InsertAt(t *testing.T) {
	t.Run("insert at beginning", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 20)
		otl.insertTx(tx1)

		tx2 := createTx([]byte("tx2"), "sender", 10)
		index := otl.findInsertionIndex(tx2)
		require.Equal(t, 0, index)

		added := otl.insertAt(tx2, index)
		require.True(t, added)
		require.Equal(t, 2, otl.len())
		require.Equal(t, tx2, otl.get(0))
		require.Equal(t, tx1, otl.get(1))
	})

	t.Run("insert at end", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 10)
		otl.insertTx(tx1)

		tx2 := createTx([]byte("tx2"), "sender", 20)
		index := otl.findInsertionIndex(tx2)
		require.Equal(t, 1, index)

		added := otl.insertAt(tx2, index)
		require.True(t, added)
		require.Equal(t, 2, otl.len())
		require.Equal(t, tx1, otl.get(0))
		require.Equal(t, tx2, otl.get(1))
	})

	t.Run("insert in middle", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 10)
		tx3 := createTx([]byte("tx3"), "sender", 30)
		otl.insertTx(tx1)
		otl.insertTx(tx3)

		tx2 := createTx([]byte("tx2"), "sender", 20)
		index := otl.findInsertionIndex(tx2)
		require.Equal(t, 1, index)

		added := otl.insertAt(tx2, index)
		require.True(t, added)
		require.Equal(t, 3, otl.len())
		require.Equal(t, tx1, otl.get(0))
		require.Equal(t, tx2, otl.get(1))
		require.Equal(t, tx3, otl.get(2))
	})

	t.Run("reject duplicate", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 10)
		otl.insertTx(tx1)

		// Same transaction (duplicate)
		tx1Dup := createTx([]byte("tx1"), "sender", 10)
		index := otl.findInsertionIndex(tx1Dup)
		require.Equal(t, 0, index)

		added := otl.insertAt(tx1Dup, index)
		require.False(t, added)
		require.Equal(t, 1, otl.len())
	})

	t.Run("insert into empty list", func(t *testing.T) {
		otl := newOrderedTransactionsList()
		tx1 := createTx([]byte("tx1"), "sender", 10)

		index := otl.findInsertionIndex(tx1)
		require.Equal(t, 0, index)

		added := otl.insertAt(tx1, index)
		require.True(t, added)
		require.Equal(t, 1, otl.len())
		require.Equal(t, tx1, otl.get(0))
	})
}

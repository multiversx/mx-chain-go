package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTxCacheFailsafe_DecoratesCache(t *testing.T) {
	backingCache := newUnconstrainedCacheToTest()
	cache := newTxCacheFailsafe(backingCache)

	// Add
	txAlice := createTx([]byte("tx-alice"), "alice", 1)
	txBob := createTx([]byte("tx-bob"), "bob", 1)
	cache.AddTx(txAlice)
	cache.AddTx(txBob)
	require.Equal(t, 2, cache.Len())

	// Get
	txAliceActual, _ := cache.GetByTxHash([]byte("tx-alice"))
	txBobActual, _ := cache.GetByTxHash([]byte("tx-bob"))
	require.Equal(t, txAlice, txAliceActual)
	require.Equal(t, txBob, txBobActual)

	// Select
	selection := cache.SelectTransactions(100, 100)
	require.ElementsMatch(t, []*WrappedTransaction{txAlice, txBob}, selection)

	// Remove
	cache.Remove([]byte("tx-bob"))
	require.Equal(t, 1, cache.Len())
}

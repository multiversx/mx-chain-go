package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDisabledCache_DoesNothing(t *testing.T) {
	cache := NewDisabledCache()

	ok, added := cache.AddTx(nil)
	require.False(t, ok)
	require.False(t, added)

	tx, ok := cache.GetByTxHash([]byte{})
	require.Nil(t, tx)
	require.False(t, ok)

	selection := cache.SelectTransactions(42, 42)
	require.Equal(t, 0, len(selection))

	err := cache.RemoveTxByHash([]byte{})
	require.Nil(t, err)

	count := cache.CountTx()
	require.Equal(t, int64(0), count)

	length := cache.Len()
	require.Equal(t, 0, length)

	require.NotPanics(t, func() { cache.ForEachTransaction(func(_ []byte, _ *WrappedTransaction) {}) })

	cache.Clear()
	evicted := cache.Put(nil, nil)
	require.False(t, evicted)
}

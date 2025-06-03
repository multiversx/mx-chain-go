package txcache

import (
	"math"
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func TestTxCache_AutoClean_Dummy(t *testing.T) {
	t.Run("with lower nonces", func(t *testing.T) {
		cache := newUnconstrainedCacheToTest()
		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 2)
		session.SetNonce([]byte("bob"), 42)
		session.SetNonce([]byte("carol"), 7)

		// Good
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3))

		// A few with lower nonce
		cache.AddTx(createTx([]byte("hash-bob-42"), "bob", 40))
		cache.AddTx(createTx([]byte("hash-bob-43"), "bob", 41))
		cache.AddTx(createTx([]byte("hash-bob-44"), "bob", 42))

		// Good
		cache.AddTx(createTx([]byte("hash-carol-7"), "carol", 7))
		cache.AddTx(createTx([]byte("hash-carol-8"), "carol", 8))

		expectedNumEvicted := 3 // 2 bob, 1 alice
		evicted:= cache.Cleanup(session, 7, math.MaxInt, selectionLoopMaximumDuration)
		require.Equal(t, uint64(expectedNumEvicted), evicted)
	})

	t.Run("with duplicated nonces", func(t *testing.T) {
		cache := newUnconstrainedCacheToTest()
		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)

		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
		cache.AddTx(createTx([]byte("hash-alice-3a"), "alice", 3))
		cache.AddTx(createTx([]byte("hash-alice-3b"), "alice", 3).withGasPrice(oneBillion * 2))
		cache.AddTx(createTx([]byte("hash-alice-3c"), "alice", 3))
		cache.AddTx(createTx([]byte("hash-alice-4"), "alice", 4))

		// Check that the duplicates are removed
		evicted:= cache.Cleanup(session, 5, math.MaxInt, selectionLoopMaximumDuration)
		require.Equal(t, uint64(2), evicted) // duplicates for nonce 3
		
		// Check that the duplicates were removed based on their lower priority
		listForAlice, _ := cache.txListBySender.getListForSender("alice")
		require.Equal(t, 4, listForAlice.items.Len())
		for element := listForAlice.items.Front(); element != nil; {
			tx := element.Value.(*WrappedTransaction)
			if tx.Tx.GetNonce() == 3 {
				require.Equal(t, oneBillion*2, int(tx.Tx.GetGasPrice()))
			} else {
				require.Equal(t, oneBillion, int(tx.Tx.GetGasPrice()))
			}
			element = element.Next()
		}

	})
}


package txcache

import (
	"math"
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxCache_SenderOrder_Dummy(t *testing.T) {
	t.Run("shuffle addresses", func(t *testing.T) {
		selectionConfig := createMockTxCacheSelectionConfig(math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(selectionConfig, boundsConfig)

		//add data into the cache
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-bob-40"), "bob", 40))
		cache.AddTx(createTx([]byte("hash-carol-7"), "carol", 7))

		sender_addresses := cache.txListBySender.backingMap.Keys()
		shuffled_addresses := shuffleSendersAddresses(sender_addresses, 108)
		same_randomness_shuffled_addresses := shuffleSendersAddresses(sender_addresses, 108)
		other_randomness_shuffled_addresses := shuffleSendersAddresses(sender_addresses, 10832)
		
		assert.NotEqual(t, sender_addresses, shuffled_addresses)
		assert.NotEqual(t, sender_addresses, other_randomness_shuffled_addresses)
		assert.NotEqual(t, shuffled_addresses, other_randomness_shuffled_addresses)
		assert.Equal(t, shuffled_addresses, same_randomness_shuffled_addresses)
	})	
}

func TestTxCache_AutoClean_Dummy(t *testing.T) {
	t.Run("with lower nonces", func(t *testing.T) {
		selectionConfig := createMockTxCacheSelectionConfig(math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(selectionConfig, boundsConfig)

		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 2)
		session.SetNonce([]byte("bob"), 42)
		session.SetNonce([]byte("carol"), 7)

		// One with lower nonce
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3))

		// A few with lower nonce
		cache.AddTx(createTx([]byte("hash-bob-40"), "bob", 40))
		cache.AddTx(createTx([]byte("hash-bob-41"), "bob", 41))
		cache.AddTx(createTx([]byte("hash-bob-42"), "bob", 42))

		// Good
		cache.AddTx(createTx([]byte("hash-carol-7"), "carol", 7))
		cache.AddTx(createTx([]byte("hash-carol-8"), "carol", 8))

		expectedNumEvicted := 3 // 2 bob, 1 alice
		evicted := cache.Cleanup(session, 7, math.MaxInt, 1000 * selectionLoopMaximumDuration)
		require.Equal(t, uint64(expectedNumEvicted), evicted)
	})

	t.Run("with duplicated nonces", func(t *testing.T) {
		selectionConfig := createMockTxCacheSelectionConfig(math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(selectionConfig, boundsConfig)

		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)

		cache.AddTx(createTx([]byte("hash-alice-3a"), "alice", 3))
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
		cache.AddTx(createTx([]byte("hash-alice-3b"), "alice", 3).withGasPrice(oneBillion * 2))
		cache.AddTx(createTx([]byte("hash-alice-3c"), "alice", 3))
		cache.AddTx(createTx([]byte("hash-alice-4"), "alice", 4))

		// Check that the duplicates are removed
		evicted:= cache.Cleanup(session, 5, math.MaxInt, 1000 * selectionLoopMaximumDuration)
		require.Equal(t, uint64(2), evicted) // duplicates for nonce 3
		
		// Check that the duplicates were removed based on their lower priority
		listForAlice, _ := cache.txListBySender.getListForSender("alice")
		require.Equal(t, 4, listForAlice.items.Len())
		
		expected := map[uint64][]byte{
			1: []byte("hash-alice-1"),
			2: []byte("hash-alice-2"),
			3: []byte("hash-alice-3b"),
			4: []byte("hash-alice-4"),
		}
		
		for element := listForAlice.items.Front(); element != nil; element = element.Next() {
			tx := element.Value.(*WrappedTransaction)
			expectedHash, ok := expected[tx.Tx.GetNonce()]
			require.True(t, ok, "Unexpected nonce %d", tx.Tx.GetNonce())
			require.Equal(t, expectedHash, tx.TxHash)
		}

	})
}


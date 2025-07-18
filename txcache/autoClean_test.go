package txcache

import (
	"math"
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxCache_ShuffleSendersAddresses_Dummy(t *testing.T) {
	
	//selectionConfig := createMockTxCacheSelectionConfig(math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
	boundsConfig := createMockTxBoundsConfig()
	cache := newUnconstrainedCacheToTest(boundsConfig)

	// add data into the cache
	cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
	cache.AddTx(createTx([]byte("hash-bob-40"), "bob", 40))
	cache.AddTx(createTx([]byte("hash-carol-7"), "carol", 7))
	
	t.Run("with same randomness", func(t *testing.T) {
		senderAddresses := cache.txListBySender.backingMap.Keys()
		shuffledAddresses := shuffleSendersAddresses(senderAddresses, 108)
		sameRandomnessShuffledAddresses := shuffleSendersAddresses(senderAddresses, 108)

		assert.NotEqual(t, senderAddresses, shuffledAddresses)
		assert.Equal(t, shuffledAddresses, sameRandomnessShuffledAddresses)
	})
	
	t.Run("with different randomness", func(t *testing.T) {
		senderAddresses := cache.txListBySender.backingMap.Keys()
		shuffledAddresses := shuffleSendersAddresses(senderAddresses, 108)
		otherRandomnessShuffledAddresses := shuffleSendersAddresses(senderAddresses, 10832)
		
		assert.NotEqual(t, senderAddresses, shuffledAddresses)
		assert.NotEqual(t, senderAddresses, otherRandomnessShuffledAddresses)
		assert.NotEqual(t, shuffledAddresses, otherRandomnessShuffledAddresses)
	})
}

func TestTxCache_GetDeterministicallyShuffledSenders_Dummy(t *testing.T) {
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)

		// add data into the cache
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-bob-40"), "bob", 40))
		cache.AddTx(createTx([]byte("hash-carol-7"), "carol", 7))

	t.Run("with same randomness", func(t *testing.T) {
		sendersList := cache.txListBySender.backingMap.Keys()
		randomnessSendersList := cache.getDeterministicallyShuffledSenders(uint64(100))
		sameRandomnessSendersList := cache.getDeterministicallyShuffledSenders(uint64(100))
		
		assert.NotEqual(t, sendersList, sameRandomnessSendersList)
		assert.NotEqual(t, sendersList, randomnessSendersList)
		assert.Equal(t, randomnessSendersList, sameRandomnessSendersList)
	})
	
	t.Run("with different randomness", func(t *testing.T) {
		sendersList := cache.txListBySender.backingMap.Keys()
		randomnessSendersList := cache.getDeterministicallyShuffledSenders(uint64(100))
		otherRandomnessSendersList := cache.getDeterministicallyShuffledSenders(uint64(127))
		
		assert.NotEqual(t, sendersList, otherRandomnessSendersList)
		assert.NotEqual(t, sendersList, randomnessSendersList)
		assert.NotEqual(t, randomnessSendersList, otherRandomnessSendersList)
	})
}

func Test_RemoveSweepableTransactionsReturnHashes_Dummy(t *testing.T) {
	t.Run("with lower nonces", func(t *testing.T) {
		list := newUnconstrainedListToTest()
		list.AddTx(createTx([]byte("a"), ".", 1))
		list.AddTx(createTx([]byte("b"), ".", 3))
		list.AddTx(createTx([]byte("c"), ".", 4))
		list.AddTx(createTx([]byte("d"), ".", 2))
		list.AddTx(createTx([]byte("e"), ".", 5))

		hashesBeforeEviction := list.getTxHashesAsStrings()
		
		list.removeSweepableTransactionsReturnHashes(uint64(3))
		hashesAfterEviction := list.getTxHashesAsStrings()
		require.Equal(t, []string{"c", "e"}, hashesAfterEviction)

		expectedEvicted := 3 // nonce 1, 2, 3
		require.Equal(t, len(hashesBeforeEviction), expectedEvicted + len(hashesAfterEviction))
	})

	t.Run("with duplicate nonces, same gas", func(t *testing.T) {
		list := newUnconstrainedListToTest()

		list.AddTx(createTx([]byte("a"), ".", 1))
		list.AddTx(createTx([]byte("b"), ".", 3))
		list.AddTx(createTx([]byte("c"), ".", 3))
		list.AddTx(createTx([]byte("d"), ".", 2))
		list.AddTx(createTx([]byte("e"), ".", 3))
	
		hashesBeforeEviction := list.getTxHashesAsStrings()
		require.Equal(t, []string{"a", "d", "b", "c", "e"}, hashesBeforeEviction)
		
		list.removeSweepableTransactionsReturnHashes(uint64(0))
		hashesAfterEviction := list.getTxHashesAsStrings()
		require.Equal(t, []string{"a", "d", "b"}, hashesAfterEviction)

		expectedEvicted := 2 // nonce 3 "c", 3 "e"
		require.Equal(t, len(hashesBeforeEviction), expectedEvicted + len(hashesAfterEviction))
	})

	t.Run("with duplicate nonces, different gas", func(t *testing.T) {
		list := newUnconstrainedListToTest()
		
		list.AddTx(createTx([]byte("a"), ".", 1).withGasPrice(oneBillion))
		list.AddTx(createTx([]byte("b"), ".", 3).withGasPrice(3 * oneBillion))
		list.AddTx(createTx([]byte("c"), ".", 3).withGasPrice(3 * oneBillion))
		list.AddTx(createTx([]byte("d"), ".", 3).withGasPrice(2 * oneBillion))
		list.AddTx(createTx([]byte("e"), ".", 3).withGasPrice(3.5 * oneBillion))
		list.AddTx(createTx([]byte("f"), ".", 2).withGasPrice(oneBillion))
		list.AddTx(createTx([]byte("g"), ".", 3).withGasPrice(2.5 * oneBillion))
	
		hashesBeforeEviction := list.getTxHashesAsStrings()
		require.Equal(t, []string{"a", "f", "e", "b", "c", "g", "d"}, hashesBeforeEviction)
		
		list.removeSweepableTransactionsReturnHashes(uint64(0))
		hashesAfterEviction := list.getTxHashesAsStrings()
		require.Equal(t, []string{"a", "f", "e"}, hashesAfterEviction)

		expectedEvicted := 4 // nonce 3 hashes "b", "c", "d"
		require.Equal(t, len(hashesBeforeEviction), expectedEvicted + len(hashesAfterEviction))
	})

	t.Run("with lower nonces and duplicate nonces", func(t *testing.T) {
		list := newUnconstrainedListToTest()

		// lower nonces
		list.AddTx(createTx([]byte("a"), ".", 1))
		list.AddTx(createTx([]byte("b"), ".", 3).withGasPrice(1.2 * oneBillion))
		list.AddTx(createTx([]byte("c"), ".", 3).withGasPrice(1.1 * oneBillion))
		list.AddTx(createTx([]byte("d"), ".", 2))
		list.AddTx(createTx([]byte("e"), ".", 3).withGasPrice(1.3 * oneBillion))

		// duplicate nonces
		list.AddTx(createTx([]byte("f"), ".", 4).withGasPrice(oneBillion))
		list.AddTx(createTx([]byte("g"), ".", 6).withGasPrice(3 * oneBillion))
		list.AddTx(createTx([]byte("h"), ".", 6).withGasPrice(3.5 * oneBillion))
		list.AddTx(createTx([]byte("i"), ".", 6).withGasPrice(2 * oneBillion))
		list.AddTx(createTx([]byte("j"), ".", 6).withGasPrice(3.5 * oneBillion))
		list.AddTx(createTx([]byte("k"), ".", 5).withGasPrice(oneBillion))
		list.AddTx(createTx([]byte("l"), ".", 6).withGasPrice(2.5 * oneBillion))
	
	
		hashesBeforeEviction := list.getTxHashesAsStrings()
		require.Equal(t, []string{"a", "d", "e", "b", "c", "f", "k", "h", "j", "g", "l", "i"}, hashesBeforeEviction)
		
		list.removeSweepableTransactionsReturnHashes(uint64(3))
		hashesAfterEviction := list.getTxHashesAsStrings()
		require.Equal(t, []string{"f", "k", "h"}, hashesAfterEviction)

		expectedEvicted := 9 // lower nonces 1-3, duplicates for nonce 6
		require.Equal(t, len(hashesBeforeEviction), expectedEvicted + len(hashesAfterEviction))
	})
}


func TestTxCache_AutoClean_Dummy(t *testing.T) {
	t.Run("with lower nonces", func(t *testing.T) {
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)

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
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)

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

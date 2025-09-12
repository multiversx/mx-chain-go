package txcache

import (
	"fmt"
	"math"
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxCache_ShuffleSendersAddresses_Dummy(t *testing.T) {

	boundsConfig := createMockTxBoundsConfig()
	cache := newUnconstrainedCacheToTest(boundsConfig)

	// add data into the cache
	cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
	cache.AddTx(createTx([]byte("hash-bob-40"), "bob", 40))
	cache.AddTx(createTx([]byte("hash-carol-7"), "carol", 7))
	cache.AddTx(createTx([]byte("hash-dave-2"), "dave", 2))
	cache.AddTx(createTx([]byte("hash-eve-3"), "eve", 3))
	cache.AddTx(createTx([]byte("hash-frank-5"), "frank", 5))
	cache.AddTx(createTx([]byte("hash-grace-6"), "grace", 6))
	cache.AddTx(createTx([]byte("hash-helen-4"), "helen", 4))

	t.Run("with same randomness", func(t *testing.T) {
		senderAddresses := cache.txListBySender.backingMap.Keys()
		senderAddressesCopy := make([]string, len(senderAddresses))
		copy(senderAddressesCopy, senderAddresses)
		senderAddressesReference := make([]string, len(senderAddresses))
		copy(senderAddressesReference, senderAddresses)

		shuffleSendersAddresses(senderAddresses, 108)
		assert.NotEqual(t, senderAddresses, senderAddressesReference)

		shuffleSendersAddresses(senderAddressesCopy, 108)
		assert.NotEqual(t, senderAddressesCopy, senderAddressesReference)
		assert.Equal(t, senderAddressesCopy, senderAddresses)
	})

	t.Run("with different randomness", func(t *testing.T) {
		senderAddresses := cache.txListBySender.backingMap.Keys()
		senderAddressesCopy := make([]string, len(senderAddresses))
		copy(senderAddressesCopy, senderAddresses)
		senderAddressesReference := make([]string, len(senderAddresses))
		copy(senderAddressesReference, senderAddresses)

		shuffleSendersAddresses(senderAddresses, 108)
		assert.NotEqual(t, senderAddresses, senderAddressesReference)

		shuffleSendersAddresses(senderAddressesCopy, 10832)
		assert.NotEqual(t, senderAddressesCopy, senderAddressesReference)
		assert.NotEqual(t, senderAddressesCopy, senderAddresses)
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
		require.Equal(t, len(hashesBeforeEviction), expectedEvicted+len(hashesAfterEviction))
	})
}

func TestTxCache_AutoClean_Dummy(t *testing.T) {
	t.Run("with GetAccountNonce errors", func(t *testing.T) {
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)
		accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMock()
		accountsProvider.GetAccountNonceCalled = func(address []byte) (uint64, bool, error) {
			switch string(address) {
			case "alice":
				return 3, true, nil
			case "bob":
				return 42, true, fmt.Errorf("forced error for address %s", address)
			case "carol":
				return 7, true, nil
			default:
				return 0, false, nil
			}
		}

		// Two with lower nonce for alice
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3))
		// Two with lower nonce for bob (all will error out)
		cache.AddTx(createTx([]byte("hash-bob-40"), "bob", 40))
		cache.AddTx(createTx([]byte("hash-bob-41"), "bob", 41))
		cache.AddTx(createTx([]byte("hash-bob-42"), "bob", 42))
		// Good for carol
		cache.AddTx(createTx([]byte("hash-carol-7"), "carol", 7))
		cache.AddTx(createTx([]byte("hash-carol-8"), "carol", 8))
		expectedNumEvicted := 2 // only alice
		evicted := cache.Cleanup(accountsProvider, 7, math.MaxInt, 1000*cleanupLoopMaximumDuration)
		require.NoError(t, nil)
		require.Equal(t, uint64(expectedNumEvicted), evicted)
	})

	t.Run("with nonce equal 0", func(t *testing.T) {
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)
		accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMock()
		accountsProvider.SetNonce([]byte("alice"), 0)

		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))

		expectedNumEvicted := 0
		evicted := cache.Cleanup(accountsProvider, 7, math.MaxInt, 1000*cleanupLoopMaximumDuration)
		require.NoError(t, nil)
		require.Equal(t, uint64(expectedNumEvicted), evicted)
	})

	t.Run("with number of evicted tranzactions cap reached", func(t *testing.T) {
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)

		accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMock()
		accountsProvider.SetNonce([]byte("alice"), 3)
		accountsProvider.SetNonce([]byte("bob"), 42)

		// Two with lower nonce for alice
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3))

		// A few with lower nonce for bob
		cache.AddTx(createTx([]byte("hash-bob-40"), "bob", 40))
		cache.AddTx(createTx([]byte("hash-bob-41"), "bob", 41))
		cache.AddTx(createTx([]byte("hash-bob-42"), "bob", 42))

		expectedNumEvicted := 2 // only alice, because maxNum is 2
		evicted := cache.Cleanup(accountsProvider, 7, 2, 1000*cleanupLoopMaximumDuration)
		require.Equal(t, uint64(expectedNumEvicted), evicted)
	})

	t.Run("with cleanupLoopMaximumDuration cap reached", func(t *testing.T) {
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)
		accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMock()
		accountsProvider.SetNonce([]byte("alice"), 4)
		accountsProvider.SetNonce([]byte("bob"), 43)
		accountsProvider.SetNonce([]byte("carol"), 9)
		// Two with lower nonce for alice
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3))
		// A few with lower nonce for bob
		cache.AddTx(createTx([]byte("hash-bob-40"), "bob", 40))
		cache.AddTx(createTx([]byte("hash-bob-41"), "bob", 41))
		cache.AddTx(createTx([]byte("hash-bob-42"), "bob", 42))
		// Good for carol
		cache.AddTx(createTx([]byte("hash-carol-7"), "carol", 7))
		cache.AddTx(createTx([]byte("hash-carol-8"), "carol", 8))
		evictable := 8
		evicted := cache.Cleanup(accountsProvider, 7, math.MaxInt, 1000)
		require.Less(t, evicted, uint64(evictable))
	})

	t.Run("with lower nonces", func(t *testing.T) {
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)

		accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMock()
		accountsProvider.SetNonce([]byte("alice"), 2)
		accountsProvider.SetNonce([]byte("bob"), 42)
		accountsProvider.SetNonce([]byte("carol"), 7)

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
		evicted := cache.Cleanup(accountsProvider, 7, math.MaxInt, 1000*cleanupLoopMaximumDuration)
		require.Equal(t, uint64(expectedNumEvicted), evicted)
	})
}

// helper function for creating a new unconstrained cache with a given size
func newTxPoolWithN(size int, accountsProvider *txcachemocks.AccountNonceAndBalanceProviderMock) *TxCache {
	boundsConfig := createMockTxBoundsConfig()
	cache := newUnconstrainedCacheToTest(boundsConfig)
	for i := 0; i < size; i++ {
		cache.AddTx(createTx([]byte(fmt.Sprintf("hash-%d", i)), fmt.Sprintf("sender-%d", i), uint64(i)))
		accountsProvider.SetNonce([]byte(fmt.Sprintf("sender-%d", i)), uint64(i+10))
	}
	return cache
}

func BenchmarkAddressShuffling(b *testing.B) {
	sizes := []int{1000, 10000, 50000, 100000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			// prepare pool
			accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMock()
			cache := newTxPoolWithN(size, accountsProvider)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				senderAddresses := cache.txListBySender.backingMap.Keys()
				shuffleSendersAddresses(senderAddresses, uint64(i))
			}
		})
	}
}

func BenchmarkCleanup(b *testing.B) {
	sizes := []int{1000, 10000, 50000, 100000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMock()
				cache := newTxPoolWithN(size, accountsProvider)
				b.StartTimer()

				_ = cache.Cleanup(accountsProvider, uint64(i), math.MaxInt, 1000*cleanupLoopMaximumDuration)
			}

		})
	}
}

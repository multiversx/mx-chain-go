package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSweeping_CollectSweepable(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	cache.AddTx(createTx([]byte("alice-42"), "alice", 42))
	cache.AddTx(createTx([]byte("bob-42"), "bob", 42))
	cache.AddTx(createTx([]byte("carol-42"), "carol", 42))

	// Senders have no initial gaps
	selection := cache.doSelectTransactions(1000, 1000)
	require.Equal(t, 3, len(selection))
	require.Equal(t, 0, len(cache.sweepingListOfSenders))

	// Alice and Bob have initial gaps, Carol doesn't
	cache.NotifyAccountNonce([]byte("alice"), 10)
	cache.NotifyAccountNonce([]byte("bob"), 20)

	// 1st fail
	selection = cache.doSelectTransactions(1000, 1000)
	require.Equal(t, 1, len(selection))
	require.Equal(t, 0, len(cache.sweepingListOfSenders))
	require.Equal(t, 1, cache.getNumFailedSelectionsOfSender("alice"))
	require.Equal(t, 1, cache.getNumFailedSelectionsOfSender("bob"))
	require.Equal(t, 0, cache.getNumFailedSelectionsOfSender("carol"))

	// 2nd fail, grace period, one grace transaction for Alice and Bob
	selection = cache.doSelectTransactions(1000, 1000)
	require.Equal(t, 3, len(selection))
	require.Equal(t, 0, len(cache.sweepingListOfSenders))
	require.Equal(t, 2, cache.getNumFailedSelectionsOfSender("alice"))
	require.Equal(t, 2, cache.getNumFailedSelectionsOfSender("bob"))
	require.Equal(t, 0, cache.getNumFailedSelectionsOfSender("carol"))

	// 3nd fail, collect Alice and Bob as sweepables
	selection = cache.doSelectTransactions(1000, 1000)
	require.Equal(t, 1, len(selection))
	require.Equal(t, 2, len(cache.sweepingListOfSenders))
	require.True(t, cache.isSenderSweepable("alice"))
	require.True(t, cache.isSenderSweepable("bob"))
	require.Equal(t, 3, cache.getNumFailedSelectionsOfSender("alice"))
	require.Equal(t, 3, cache.getNumFailedSelectionsOfSender("bob"))
	require.Equal(t, 0, cache.getNumFailedSelectionsOfSender("carol"))
}

func TestSweeping_WhenSendersEscapeCollection(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	cache.AddTx(createTx([]byte("alice-42"), "alice", 42))
	cache.AddTx(createTx([]byte("bob-42"), "bob", 42))
	cache.AddTx(createTx([]byte("carol-42"), "carol", 42))

	// Senders have no initial gaps
	selection := cache.doSelectTransactions(1000, 1000)
	require.Equal(t, 3, len(selection))
	require.Equal(t, 0, len(cache.sweepingListOfSenders))

	// Alice and Bob have initial gaps, Carol doesn't
	cache.NotifyAccountNonce([]byte("alice"), 10)
	cache.NotifyAccountNonce([]byte("bob"), 20)

	// 1st fail
	selection = cache.doSelectTransactions(1000, 1000)
	require.Equal(t, 1, len(selection))
	require.Equal(t, 0, len(cache.sweepingListOfSenders))
	require.Equal(t, 1, cache.getNumFailedSelectionsOfSender("alice"))
	require.Equal(t, 1, cache.getNumFailedSelectionsOfSender("bob"))
	require.Equal(t, 0, cache.getNumFailedSelectionsOfSender("carol"))

	// 2nd fail, grace period, one grace transaction for Alice and Bob
	selection = cache.doSelectTransactions(1000, 1000)
	require.Equal(t, 3, len(selection))
	require.Equal(t, 0, len(cache.sweepingListOfSenders))
	require.Equal(t, 2, cache.getNumFailedSelectionsOfSender("alice"))
	require.Equal(t, 2, cache.getNumFailedSelectionsOfSender("bob"))
	require.Equal(t, 0, cache.getNumFailedSelectionsOfSender("carol"))

	// 3rd attempt, but with gaps resolved
	// Alice and Bob escape and won't be collected as sweepables
	cache.NotifyAccountNonce([]byte("alice"), 42)
	cache.NotifyAccountNonce([]byte("bob"), 42)

	selection = cache.doSelectTransactions(1000, 1000)
	require.Equal(t, 3, len(selection))
	require.Equal(t, 0, len(cache.sweepingListOfSenders))
	require.Equal(t, 0, cache.getNumFailedSelectionsOfSender("alice"))
	require.Equal(t, 0, cache.getNumFailedSelectionsOfSender("bob"))
	require.Equal(t, 0, cache.getNumFailedSelectionsOfSender("carol"))
}

func TestSweeping_SweepSweepable(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	cache.AddTx(createTx([]byte("alice-42"), "alice", 42))
	cache.AddTx(createTx([]byte("bob-42"), "bob", 42))
	cache.AddTx(createTx([]byte("carol-42"), "carol", 42))

	// Fake "Alice" and "Bob" as sweepable
	cache.sweepingListOfSenders = []*txListForSender{
		cache.getListForSender("alice"),
		cache.getListForSender("bob"),
	}

	require.Equal(t, uint64(3), cache.CountTx())
	require.Equal(t, uint64(3), cache.CountSenders())

	cache.sweepSweepable()

	require.Equal(t, uint64(1), cache.CountTx())
	require.Equal(t, uint64(1), cache.CountSenders())
}

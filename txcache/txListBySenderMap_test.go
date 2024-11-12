package txcache

import (
	"math"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendersMap_AddTx_IncrementsCounter(t *testing.T) {
	myMap := newSendersMapToTest()

	myMap.addTxReturnEvicted(createTx([]byte("a"), "alice", 1))
	myMap.addTxReturnEvicted(createTx([]byte("aa"), "alice", 2))
	myMap.addTxReturnEvicted(createTx([]byte("b"), "bob", 1))

	// There are 2 senders
	require.Equal(t, int64(2), myMap.counter.Get())
}

func TestSendersMap_removeTransactionsWithLowerOrEqualNonceReturnHashes_alsoRemovesSenderWhenNoTransactionLeft(t *testing.T) {
	myMap := newSendersMapToTest()

	txAlice1 := createTx([]byte("a1"), "alice", 1)
	txAlice2 := createTx([]byte("a2"), "alice", 2)
	txBob := createTx([]byte("b"), "bob", 1)

	myMap.addTxReturnEvicted(txAlice1)
	myMap.addTxReturnEvicted(txAlice2)
	myMap.addTxReturnEvicted(txBob)
	require.Equal(t, int64(2), myMap.counter.Get())
	require.Equal(t, uint64(2), myMap.testGetListForSender("alice").countTx())
	require.Equal(t, uint64(1), myMap.testGetListForSender("bob").countTx())

	_ = myMap.removeTransactionsWithLowerOrEqualNonceReturnHashes(txAlice1)
	require.Equal(t, int64(2), myMap.counter.Get())
	require.Equal(t, uint64(1), myMap.testGetListForSender("alice").countTx())
	require.Equal(t, uint64(1), myMap.testGetListForSender("bob").countTx())

	_ = myMap.removeTransactionsWithLowerOrEqualNonceReturnHashes(txAlice2)
	// All alice's transactions have been removed now
	require.Equal(t, int64(1), myMap.counter.Get())

	_ = myMap.removeTransactionsWithLowerOrEqualNonceReturnHashes(txBob)
	// Also Bob has no more transactions
	require.Equal(t, int64(0), myMap.counter.Get())
}

func TestSendersMap_RemoveSender(t *testing.T) {
	myMap := newSendersMapToTest()

	myMap.addTxReturnEvicted(createTx([]byte("a"), "alice", 1))
	require.Equal(t, int64(1), myMap.counter.Get())

	// Bob is unknown
	myMap.removeSender("bob")
	require.Equal(t, int64(1), myMap.counter.Get())

	myMap.removeSender("alice")
	require.Equal(t, int64(0), myMap.counter.Get())
}

func TestSendersMap_RemoveSendersBulk_ConcurrentWithAddition(t *testing.T) {
	myMap := newSendersMapToTest()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for i := 0; i < 100; i++ {
			numRemoved := myMap.RemoveSendersBulk([]string{"alice"})
			require.LessOrEqual(t, numRemoved, uint32(1))

			numRemoved = myMap.RemoveSendersBulk([]string{"bob"})
			require.LessOrEqual(t, numRemoved, uint32(1))

			numRemoved = myMap.RemoveSendersBulk([]string{"carol"})
			require.LessOrEqual(t, numRemoved, uint32(1))
		}
	}()

	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func(i int) {
			myMap.addTxReturnEvicted(createTx([]byte("a"), "alice", uint64(i)))
			myMap.addTxReturnEvicted(createTx([]byte("b"), "bob", uint64(i)))
			myMap.addTxReturnEvicted(createTx([]byte("c"), "carol", uint64(i)))

			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestSendersMap_notifyAccountNonce(t *testing.T) {
	myMap := newSendersMapToTest()

	// Discarded notification, since sender not added yet
	myMap.notifyAccountNonce([]byte("alice"), 42)

	_, _ = myMap.addTxReturnEvicted(createTx([]byte("tx-42"), "alice", 42))
	alice, _ := myMap.getListForSender("alice")
	require.Equal(t, uint64(0), alice.accountNonce.Get())
	require.False(t, alice.accountNonceKnown.IsSet())

	myMap.notifyAccountNonce([]byte("alice"), 42)
	require.Equal(t, uint64(42), alice.accountNonce.Get())
	require.True(t, alice.accountNonceKnown.IsSet())
}

func newSendersMapToTest() *txListBySenderMap {
	return newTxListBySenderMap(4, senderConstraints{
		maxNumBytes: math.MaxUint32,
		maxNumTxs:   math.MaxUint32,
	})
}

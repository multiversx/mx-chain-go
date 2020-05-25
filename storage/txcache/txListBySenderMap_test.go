package txcache

import (
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendersMap_AddTx_IncrementsCounter(t *testing.T) {
	myMap := newSendersMapToTest()

	myMap.addTx(createTx([]byte("a"), "alice", uint64(1)))
	myMap.addTx(createTx([]byte("aa"), "alice", uint64(2)))
	myMap.addTx(createTx([]byte("b"), "bob", uint64(1)))

	// There are 2 senders
	require.Equal(t, int64(2), myMap.counter.Get())
}

func TestSendersMap_RemoveTx_AlsoRemovesSenderWhenNoTransactionLeft(t *testing.T) {
	myMap := newSendersMapToTest()

	txAlice1 := createTx([]byte("a1"), "alice", uint64(1))
	txAlice2 := createTx([]byte("a2"), "alice", uint64(2))
	txBob := createTx([]byte("b"), "bob", uint64(1))

	myMap.addTx(txAlice1)
	myMap.addTx(txAlice2)
	myMap.addTx(txBob)
	require.Equal(t, int64(2), myMap.counter.Get())
	require.Equal(t, uint64(2), myMap.testGetListForSender("alice").countTx())
	require.Equal(t, uint64(1), myMap.testGetListForSender("bob").countTx())

	myMap.removeTx(txAlice1)
	require.Equal(t, int64(2), myMap.counter.Get())
	require.Equal(t, uint64(1), myMap.testGetListForSender("alice").countTx())
	require.Equal(t, uint64(1), myMap.testGetListForSender("bob").countTx())

	myMap.removeTx(txAlice2)
	// All alice's transactions have been removed now
	require.Equal(t, int64(1), myMap.counter.Get())

	myMap.removeTx(txBob)
	// Also Bob has no more transactions
	require.Equal(t, int64(0), myMap.counter.Get())
}

func TestSendersMap_RemoveSender(t *testing.T) {
	myMap := newSendersMapToTest()

	myMap.addTx(createTx([]byte("a"), "alice", uint64(1)))
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
			myMap.addTx(createTx([]byte("a"), "alice", uint64(i)))
			myMap.addTx(createTx([]byte("b"), "bob", uint64(i)))
			myMap.addTx(createTx([]byte("c"), "carol", uint64(i)))

			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestSendersMap_notifyAccountNonce(t *testing.T) {
	myMap := newSendersMapToTest()

	// Discarded notification, since sender not added yet
	myMap.notifyAccountNonce([]byte("alice"), 42)

	myMap.addTx(createTx([]byte("tx-42"), "alice", uint64(42)))
	alice, _ := myMap.getListForSender("alice")
	require.Equal(t, uint64(0), alice.accountNonce.Get())
	require.False(t, alice.accountNonceKnown.IsSet())

	myMap.notifyAccountNonce([]byte("alice"), 42)
	require.Equal(t, uint64(42), alice.accountNonce.Get())
	require.True(t, alice.accountNonceKnown.IsSet())
}

func BenchmarkSendersMap_GetSnapshotAscending(b *testing.B) {
	if b.N > 10 {
		fmt.Println("impractical benchmark: b.N too high")
		return
	}

	numSenders := 250000
	maps := make([]*txListBySenderMap, b.N)
	for i := 0; i < b.N; i++ {
		maps[i] = createTxListBySenderMap(numSenders)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		measureWithStopWatch(b, func() {
			snapshot := maps[i].getSnapshotAscending()
			require.Len(b, snapshot, numSenders)
		})
	}
}

func TestSendersMap_GetSnapshots_NoPanic_IfAlsoConcurrentMutation(t *testing.T) {
	myMap := newSendersMapToTest()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(2)

		go func() {
			for j := 0; j < 100; j++ {
				myMap.getSnapshotAscending()
			}

			wg.Done()
		}()

		go func() {
			for j := 0; j < 1000; j++ {
				sender := fmt.Sprintf("Sender-%d", j)
				myMap.removeSender(sender)
			}

			wg.Done()
		}()
	}

	wg.Wait()
}

func createTxListBySenderMap(numSenders int) *txListBySenderMap {
	myMap := newSendersMapToTest()
	for i := 0; i < numSenders; i++ {
		sender := fmt.Sprintf("Sender-%d", i)
		hash := createFakeTxHash([]byte(sender), 1)
		myMap.addTx(createTx(hash, sender, uint64(1)))
	}

	return myMap
}

func newSendersMapToTest() *txListBySenderMap {
	return newTxListBySenderMap(4, senderConstraints{
		maxNumBytes: math.MaxUint32,
		maxNumTxs:   math.MaxUint32,
	}, &disabledScoreComputer{})
}

package txcache

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_AddTx_IncrementsCounter(t *testing.T) {
	myMap := newSendersMapToTest()

	myMap.addTx([]byte("a"), createTx("alice", uint64(1)))
	myMap.addTx([]byte("aa"), createTx("alice", uint64(2)))
	myMap.addTx([]byte("b"), createTx("bob", uint64(1)))

	// There are 2 senders
	require.Equal(t, int64(2), myMap.counter.Get())
}

func Test_RemoveTx_AlsoRemovesSenderWhenNoTransactionLeft(t *testing.T) {
	myMap := newSendersMapToTest()

	txAlice1 := createTx("alice", uint64(1))
	txAlice2 := createTx("alice", uint64(2))
	txBob := createTx("bob", uint64(1))

	myMap.addTx([]byte("a"), txAlice1)
	myMap.addTx([]byte("a"), txAlice2)
	myMap.addTx([]byte("b"), txBob)
	require.Equal(t, int64(2), myMap.counter.Get())

	myMap.removeTx(txAlice1)
	require.Equal(t, int64(2), myMap.counter.Get())

	myMap.removeTx(txAlice2)
	// All alice's transactions have been removed now
	require.Equal(t, int64(1), myMap.counter.Get())

	myMap.removeTx(txBob)
	// Also Bob has no more transactions
	require.Equal(t, int64(0), myMap.counter.Get())
}

func Test_RemoveSender(t *testing.T) {
	myMap := newSendersMapToTest()

	myMap.addTx([]byte("a"), createTx("alice", uint64(1)))
	require.Equal(t, int64(1), myMap.counter.Get())

	// Bob is unknown
	myMap.removeSender("bob")
	require.Equal(t, int64(1), myMap.counter.Get())

	myMap.removeSender("alice")
	require.Equal(t, int64(0), myMap.counter.Get())
}

func Benchmark_GetSnapshotAscending(b *testing.B) {
	if b.N > 10 {
		fmt.Println("impractical benchmark: b.N too high")
		return
	}

	numSenders := 250000
	maps := make([]txListBySenderMap, b.N)
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

func Test_GetSnapshots_NoPanic_IfAlsoConcurrentMutation(t *testing.T) {
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

func createTxListBySenderMap(numSenders int) txListBySenderMap {
	myMap := newSendersMapToTest()
	for i := 0; i < numSenders; i++ {
		sender := fmt.Sprintf("Sender-%d", i)
		hash := createFakeTxHash([]byte(sender), 1)
		myMap.addTx(hash, createTx(sender, uint64(1)))
	}

	return myMap
}

func newSendersMapToTest() txListBySenderMap {
	return newTxListBySenderMap(4, &CacheConfig{})
}

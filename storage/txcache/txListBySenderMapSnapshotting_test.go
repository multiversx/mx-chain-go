package txcache

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendersMap_GetSnapshotHeadSize(t *testing.T) {
	// With empty snapshot
	snapshot := makeTestSnapshot(0)
	headSize := getSnapshotHeadSize(snapshot)
	require.Equal(t, 0, headSize)

	// With len(snapshot) smaller < "defaultSendersSnapshotHeadSize"
	snapshot = makeTestSnapshot(42)
	headSize = getSnapshotHeadSize(snapshot)
	require.Equal(t, 42, headSize)

	// With len(snapshot) == "defaultSendersSnapshotHeadSize"
	snapshot = makeTestSnapshot(defaultSendersSnapshotHeadSize)
	headSize = getSnapshotHeadSize(snapshot)
	require.Equal(t, defaultSendersSnapshotHeadSize, headSize)

	// With len(snapshot) > "defaultSendersSnapshotHeadSize", and ceiling needed
	// All but the last 43 have score 0
	snapshot = makeTestSnapshot(defaultSendersSnapshotHeadSize + 42 + 43)
	setScoreToTestSnapshotInterval(snapshot, defaultSendersSnapshotHeadSize+42, 43, 1)
	headSize = getSnapshotHeadSize(snapshot)
	require.Equal(t, defaultSendersSnapshotHeadSize+42, headSize)

	// With len(snapshot) > "defaultSendersSnapshotHeadSize", and ceiling not needed
	snapshot = makeTestSnapshot(defaultSendersSnapshotHeadSize + 42 + 43)
	setScoreToTestSnapshotInterval(snapshot, defaultSendersSnapshotHeadSize, 42+43, 1)
	headSize = getSnapshotHeadSize(snapshot)
	require.Equal(t, defaultSendersSnapshotHeadSize, headSize)
}

func TestSendersMap_GetSnapshotWithDeterministicallySortedParts(t *testing.T) {
	myMap := newSendersMapToTest()

	snapshotAscending := myMap.getSnapshotAscendingWithDeterministicallySortedHead()
	snapshotDescending := myMap.getSnapshotDescendingWithDeterministicallySortedTail()
	require.Equal(t, 0, len(snapshotAscending))
	require.Equal(t, 0, len(snapshotDescending))

	myMap.addTx(createTx([]byte("a"), "alice", uint64(1)))
	myMap.addTx(createTx([]byte("b"), "bob", uint64(1)))
	myMap.addTx(createTx([]byte("c"), "carol", uint64(1)))

	snapshotAscending = myMap.getSnapshotAscendingWithDeterministicallySortedHead()
	snapshotDescending = myMap.getSnapshotDescendingWithDeterministicallySortedTail()
	require.Equal(t, 3, len(snapshotAscending))
	require.Equal(t, 3, len(snapshotDescending))
}

func BenchmarkSendersMap_GetSnapshotAscending(b *testing.B) {
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

func makeTestSnapshot(size int) []*txListForSender {
	snapshot := make([]*txListForSender, size)

	for i := 0; i < size; i++ {
		snapshot[i] = newListToTest()
	}

	return snapshot
}

func setScoreToTestSnapshotInterval(snapshot []*txListForSender, from int, count int, score uint32) {
	for i := from; i < from+count; i++ {
		snapshot[i].lastComputedScore.Set(score)
	}
}

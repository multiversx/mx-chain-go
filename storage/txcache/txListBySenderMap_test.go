package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_AddTx_IncrementsCounter(t *testing.T) {
	myMap := newTxListBySenderMap(4)

	myMap.addTx([]byte("a"), createTx("alice", uint64(1)))
	myMap.addTx([]byte("aa"), createTx("alice", uint64(2)))
	myMap.addTx([]byte("b"), createTx("bob", uint64(1)))

	// There are 2 senders
	require.Equal(t, int64(2), myMap.counter.Get())
}

func Test_RemoveTx_AlsoRemovesSenderWhenNoTransactionLeft(t *testing.T) {
	myMap := newTxListBySenderMap(4)

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
	myMap := newTxListBySenderMap(1)

	myMap.addTx([]byte("a"), createTx("alice", uint64(1)))
	require.Equal(t, int64(1), myMap.counter.Get())

	// Bob is unknown
	myMap.removeSender("bob")
	require.Equal(t, int64(1), myMap.counter.Get())

	myMap.removeSender("alice")
	require.Equal(t, int64(0), myMap.counter.Get())
}

func Test_GetListsSortedByOrderNumber(t *testing.T) {
	myMap := newTxListBySenderMap(4)

	myMap.addTx([]byte("a"), createTx("alice", uint64(1)))
	myMap.addTx([]byte("aa"), createTx("alice", uint64(2)))
	myMap.addTx([]byte("b"), createTx("bob", uint64(1)))
	myMap.addTx([]byte("aaa"), createTx("alice", uint64(2)))
	myMap.addTx([]byte("c"), createTx("carol", uint64(2)))

	lists := myMap.GetListsSortedByOrderNumber()

	require.ElementsMatch(t, myMap.GetListsSortedByOrderNumber(), myMap.GetListsSortedBy(SortByOrderNumberAsc))
	require.ElementsMatch(t, myMap.GetListsSortedByOrderNumber(), myMap.GetListsSortedBy("foobar"))
	require.Equal(t, "alice", lists[0].sender)
	require.Equal(t, "bob", lists[1].sender)
	require.Equal(t, "carol", lists[2].sender)
}

func Test_GetListsSortedByTotalBytes(t *testing.T) {
	myMap := newTxListBySenderMap(4)

	myMap.addTx([]byte("a"), createTxWithData("alice", uint64(1), 500))
	myMap.addTx([]byte("aa"), createTxWithData("alice", uint64(2), 500))
	myMap.addTx([]byte("b"), createTxWithData("bob", uint64(1), 2000))
	myMap.addTx([]byte("aaa"), createTxWithData("alice", uint64(2), 500))
	myMap.addTx([]byte("c"), createTxWithData("carol", uint64(2), 1499))

	lists := myMap.GetListsSortedByTotalBytes()

	require.ElementsMatch(t, myMap.GetListsSortedByTotalBytes(), myMap.GetListsSortedBy(SortByTotalBytesDesc))
	require.Equal(t, "bob", lists[0].sender)
	require.Equal(t, "alice", lists[1].sender)
	require.Equal(t, "carol", lists[2].sender)
}

func Test_GetListsSortedByTotalGas(t *testing.T) {
	myMap := newTxListBySenderMap(4)

	myMap.addTx([]byte("a"), createTxWithGas("alice", uint64(1), 500, 10))
	myMap.addTx([]byte("aa"), createTxWithGas("alice", uint64(2), 500, 10))
	myMap.addTx([]byte("b"), createTxWithGas("bob", uint64(1), 500, 20))
	myMap.addTx([]byte("bb"), createTxWithGas("bob", uint64(1), 500, 20))
	myMap.addTx([]byte("c"), createTxWithGas("carol", uint64(2), 500, 15))
	myMap.addTx([]byte("cc"), createTxWithGas("carol", uint64(2), 500, 15))

	lists := myMap.GetListsSortedByTotalGas()

	require.ElementsMatch(t, myMap.GetListsSortedByTotalGas(), myMap.GetListsSortedBy(SortByTotalGas))
	require.Equal(t, "alice", lists[0].sender)
	require.Equal(t, "carol", lists[1].sender)
	require.Equal(t, "bob", lists[2].sender)
}

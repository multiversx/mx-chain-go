package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSenderAsBucketSortedMapItem_ComputeScore(t *testing.T) {
	list := newUnconstrainedListToTest()

	list.AddTx(createTxWithParams([]byte("a"), ".", 1, 1000, 200000, 100*oneBillion))
	list.AddTx(createTxWithParams([]byte("b"), ".", 1, 500, 100000, 100*oneBillion))
	list.AddTx(createTxWithParams([]byte("c"), ".", 1, 500, 100000, 100*oneBillion))

	require.Equal(t, uint64(3), list.countTx())
	require.Equal(t, int64(2000), list.totalBytes.Get())
	require.Equal(t, int64(400000), list.totalGas.Get())
	require.Equal(t, int64(40*oneMilion), list.totalFee.Get())

	require.InDelta(t, float64(5.795382396), list.computeRawScore(), delta)
}

func TestSenderAsBucketSortedMapItem_ScoreFluctuatesDeterministicallyWhenTransactionsAreAddedOrRemoved(t *testing.T) {
	list := newUnconstrainedListToTest()

	A := createTxWithParams([]byte("A"), ".", 1, 1000, 200000, 100*oneBillion)
	B := createTxWithParams([]byte("b"), ".", 1, 500, 100000, 100*oneBillion)
	C := createTxWithParams([]byte("c"), ".", 1, 500, 100000, 100*oneBillion)

	scoreNone := int(list.ComputeScore())
	list.AddTx(A)
	scoreA := int(list.ComputeScore())
	list.AddTx(B)
	scoreAB := int(list.ComputeScore())
	list.AddTx(C)
	scoreABC := int(list.ComputeScore())

	require.Equal(t, 0, scoreNone)
	require.Equal(t, 17, scoreA)
	require.Equal(t, 8, scoreAB)
	require.Equal(t, 5, scoreABC)

	list.RemoveTx(C)
	scoreAB = int(list.ComputeScore())
	list.RemoveTx(B)
	scoreA = int(list.ComputeScore())
	list.RemoveTx(A)
	scoreNone = int(list.ComputeScore())

	require.Equal(t, 0, scoreNone)
	require.Equal(t, 17, scoreA)
	require.Equal(t, 8, scoreAB)
	require.Equal(t, 5, scoreABC)
}

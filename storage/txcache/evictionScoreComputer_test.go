package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Scores_Empty(t *testing.T) {
	lists := make([]*txListForSender, 0)
	computer := newEvictionScoreComputer(lists)

	require.Equal(t, 0, len(computer.scores))
}

func Test_ParameterRanges(t *testing.T) {
	lists := make([]*txListForSender, 0)

	list := newTxListForSender("alice", 2)
	list.AddTx([]byte("alice-1"), createTxWithGas("alice", 1, 200, 15))
	list.AddTx([]byte("alice-2"), createTxWithGas("alice", 2, 200, 15))
	lists = append(lists, list)

	list = newTxListForSender("bob", 4)
	list.AddTx([]byte("bob-1"), createTxWithGas("bob", 1, 1000, 40))
	list.AddTx([]byte("bob-2"), createTxWithGas("bob", 2, 1000, 40))
	lists = append(lists, list)

	list = newTxListForSender("carol", 6)
	list.AddTx([]byte("carol-1"), createTxWithGas("carol", 1, 500, 20))
	lists = append(lists, list)

	computer := newEvictionScoreComputer(lists)

	require.Equal(t, int64(20), computer.minGas)
	require.Equal(t, int64(80), computer.maxGas)
	require.Equal(t, int64(60), computer.gasRange)

	require.Equal(t, int64(400), computer.minSize)
	require.Equal(t, int64(2000), computer.maxSize)
	require.Equal(t, int64(1600), computer.sizeRange)

	require.Equal(t, int64(1), computer.minTxCount)
	require.Equal(t, int64(2), computer.maxTxCount)
	require.Equal(t, int64(1), computer.txCountRange)

	require.Equal(t, int64(2), computer.minOrderNumber)
	require.Equal(t, int64(6), computer.maxOrderNumber)
	require.Equal(t, int64(4), computer.orderNumberRange)
}

func Test_Scores_WithRespectToOrderNumber(t *testing.T) {
	// We keep all parameters constant, except order number

	lists := make([]*txListForSender, 0)

	list := newTxListForSender("alice", 2)
	list.AddTx([]byte("alice-1"), createTxWithData("alice", 1, 200))
	list.AddTx([]byte("alice-2"), createTxWithData("alice", 2, 200))
	lists = append(lists, list)

	list = newTxListForSender("bob", 4)
	list.AddTx([]byte("bob-1"), createTxWithData("bob", 1, 200))
	list.AddTx([]byte("bob-2"), createTxWithData("bob", 2, 200))
	lists = append(lists, list)

	list = newTxListForSender("carol", 6)
	list.AddTx([]byte("carol-1"), createTxWithData("carol", 1, 200))
	list.AddTx([]byte("carol-2"), createTxWithData("carol", 2, 200))
	lists = append(lists, list)

	computer := newEvictionScoreComputer(lists)

	require.EqualValues(t, []float64{1, 1.5, 2}, computer.scores)
	require.EqualValues(t, []int64{0, 50, 100}, computer.scoresAsPercents)
}

func Test_Scores_WithRespectToSize(t *testing.T) {
	// We keep all parameters constant, except order number and size

	lists := make([]*txListForSender, 0)

	list := newTxListForSender("alice", 2)
	list.AddTx([]byte("alice-1"), createTxWithData("alice", 1, 200))
	list.AddTx([]byte("alice-2"), createTxWithData("alice", 2, 200))
	lists = append(lists, list)

	list = newTxListForSender("bob", 4)
	list.AddTx([]byte("bob-1"), createTxWithData("bob", 1, 800))
	list.AddTx([]byte("bob-2"), createTxWithData("bob", 2, 800))
	lists = append(lists, list)

	list = newTxListForSender("carol", 6)
	list.AddTx([]byte("carol-1"), createTxWithData("carol", 1, 500))
	list.AddTx([]byte("carol-2"), createTxWithData("carol", 2, 500))
	lists = append(lists, list)

	computer := newEvictionScoreComputer(lists)

	require.EqualValues(t, []float64{1, 0.75, 1.3333333333333333}, computer.scores)
	require.EqualValues(t, []int64{42, 0, 100}, computer.scoresAsPercents)
}

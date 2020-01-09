package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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

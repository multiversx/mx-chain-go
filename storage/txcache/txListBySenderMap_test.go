package txcache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AddTx_IncrementsCounter(t *testing.T) {
	myMap := NewTxListBySenderMap(100, 10)

	myMap.AddTx([]byte("a"), createTx("alice", uint64(1)))
	myMap.AddTx([]byte("aa"), createTx("alice", uint64(2)))
	myMap.AddTx([]byte("b"), createTx("bob", uint64(1)))

	// There are 2 senders
	assert.Equal(t, int64(2), myMap.Counter.Get())
}

func Test_RemoveTx_AlsoRemovesSenderWhenNoTransactionLeft(t *testing.T) {
	myMap := NewTxListBySenderMap(100, 10)

	txAlice1 := createTx("alice", uint64(1))
	txAlice2 := createTx("alice", uint64(2))
	txBob := createTx("bob", uint64(1))

	myMap.AddTx([]byte("a"), txAlice1)
	myMap.AddTx([]byte("a"), txAlice2)
	myMap.AddTx([]byte("b"), txBob)
	assert.Equal(t, int64(2), myMap.Counter.Get())

	myMap.RemoveTx(txAlice1)
	assert.Equal(t, int64(2), myMap.Counter.Get())

	myMap.RemoveTx(txAlice2)
	assert.Equal(t, int64(1), myMap.Counter.Get())

	myMap.RemoveTx(txBob)
	assert.Equal(t, int64(0), myMap.Counter.Get())
}

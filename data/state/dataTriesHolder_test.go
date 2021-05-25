package state_test

import (
	"strconv"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewDataTriesHolder(t *testing.T) {
	t.Parallel()

	dth := state.NewDataTriesHolder()
	assert.False(t, check.IfNil(dth))
}

func TestDataTriesHolder_PutAndGet(t *testing.T) {
	t.Parallel()

	tr1 := &testscommon.TrieStub{}

	dth := state.NewDataTriesHolder()
	dth.Put([]byte("trie1"), tr1)
	tr := dth.Get([]byte("trie1"))

	assert.True(t, tr == tr1)
}

func TestDataTriesHolder_GetAll(t *testing.T) {
	t.Parallel()

	tr1 := &testscommon.TrieStub{}
	tr2 := &testscommon.TrieStub{}
	tr3 := &testscommon.TrieStub{}

	dth := state.NewDataTriesHolder()
	dth.Put([]byte("trie1"), tr1)
	dth.Put([]byte("trie2"), tr2)
	dth.Put([]byte("trie3"), tr3)
	tries := dth.GetAll()

	assert.Equal(t, 3, len(tries))
}

func TestDataTriesHolder_Reset(t *testing.T) {
	t.Parallel()

	tr1 := &testscommon.TrieStub{}

	dth := state.NewDataTriesHolder()
	dth.Put([]byte("trie1"), tr1)
	dth.Reset()

	tr := dth.Get([]byte("trie1"))
	assert.Nil(t, tr)
}

func TestDataTriesHolder_Concurrency(t *testing.T) {
	t.Parallel()

	dth := state.NewDataTriesHolder()
	numTries := 50

	wg := sync.WaitGroup{}
	wg.Add(numTries)

	for i := 0; i < numTries; i++ {
		go func(key int) {
			dth.Put([]byte(strconv.Itoa(key)), &testscommon.TrieStub{})
			wg.Done()
		}(i)
	}

	wg.Wait()

	tries := dth.GetAll()
	assert.Equal(t, numTries, len(tries))
}

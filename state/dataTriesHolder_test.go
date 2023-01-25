package state_test

import (
	"strconv"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/state"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewDataTriesHolder(t *testing.T) {
	t.Parallel()

	dth := state.NewDataTriesHolder()
	assert.False(t, check.IfNil(dth))
}

func TestDataTriesHolder_PutAndGet(t *testing.T) {
	t.Parallel()

	tr1 := &trieMock.TrieStub{}

	dth := state.NewDataTriesHolder()
	dth.Put([]byte("trie1"), tr1)
	tr := dth.Get([]byte("trie1"))

	assert.True(t, tr == tr1)
}

func TestDataTriesHolder_Replace(t *testing.T) {
	t.Parallel()

	tr1 := &trieMock.TrieStub{}
	tr2 := &trieMock.TrieStub{}

	dth := state.NewDataTriesHolder()
	dth.Put([]byte("trie1"), tr1)
	dth.Replace([]byte("trie1"), tr2)
	retrievedTrie := dth.Get([]byte("trie1"))

	assert.True(t, retrievedTrie == tr2)
	assert.True(t, retrievedTrie != tr1)
}

func TestDataTriesHolder_GetAll(t *testing.T) {
	t.Parallel()

	tr1 := &trieMock.TrieStub{}
	tr2 := &trieMock.TrieStub{}
	tr3 := &trieMock.TrieStub{}

	dth := state.NewDataTriesHolder()
	dth.Put([]byte("trie1"), tr1)
	dth.Put([]byte("trie2"), tr2)
	dth.Put([]byte("trie3"), tr3)
	tries := dth.GetAll()

	assert.Equal(t, 3, len(tries))
}

func TestDataTriesHolder_Reset(t *testing.T) {
	t.Parallel()

	tr1 := &trieMock.TrieStub{}

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
			dth.Put([]byte(strconv.Itoa(key)), &trieMock.TrieStub{})
			wg.Done()
		}(i)
	}

	wg.Wait()

	tries := dth.GetAll()
	assert.Equal(t, numTries, len(tries))
}

func TestDataTriesHolder_GetAllTries(t *testing.T) {
	t.Parallel()

	dth := state.NewDataTriesHolder()
	numTries := 50

	wg := sync.WaitGroup{}
	wg.Add(numTries)

	for i := 0; i < numTries; i++ {
		go func(key int) {
			dth.Put([]byte(strconv.Itoa(key)), &trieMock.TrieStub{})
			wg.Done()
		}(i)
	}

	wg.Wait()

	tries := dth.GetAllTries()
	assert.Equal(t, numTries, len(tries))
}

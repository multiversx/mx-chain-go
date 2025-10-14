package triesHolder

import (
	"strconv"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewTriesHolder(t *testing.T) {
	t.Parallel()

	dth := NewTriesHolder()
	assert.False(t, check.IfNil(dth))
}

func TestTriesHolder_PutAndGet(t *testing.T) {
	t.Parallel()

	tr1 := &trieMock.TrieStub{}

	dth := NewTriesHolder()
	dth.Put([]byte("trie1"), tr1)
	tr := dth.Get([]byte("trie1"))

	assert.True(t, tr == tr1)
}

func TestTriesHolder_GetAll(t *testing.T) {
	t.Parallel()

	tr1 := &trieMock.TrieStub{}
	tr2 := &trieMock.TrieStub{}
	tr3 := &trieMock.TrieStub{}

	dth := NewTriesHolder()
	dth.Put([]byte("trie1"), tr1)
	dth.Put([]byte("trie2"), tr2)
	dth.Put([]byte("trie3"), tr3)
	tries := dth.GetAll()

	assert.Equal(t, 3, len(tries))
}

func TestTriesHolder_Reset(t *testing.T) {
	t.Parallel()

	tr1 := &trieMock.TrieStub{}

	dth := NewTriesHolder()
	dth.Put([]byte("trie1"), tr1)
	dth.Reset()

	tr := dth.Get([]byte("trie1"))
	assert.Nil(t, tr)
}

func TestTriesHolder_Concurrency(t *testing.T) {
	t.Parallel()

	dth := NewTriesHolder()
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

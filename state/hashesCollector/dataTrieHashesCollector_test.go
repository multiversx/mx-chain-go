package hashesCollector

import (
	"strconv"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewDataTrieHashesCollector(t *testing.T) {
	t.Parallel()

	hc := NewDataTrieHashesCollector()
	assert.False(t, check.IfNil(hc))
	assert.NotNil(t, hc.oldHashes)
	assert.NotNil(t, hc.newHashes)
}

func TestDataTrieHashesCollector_AddDirtyHash(t *testing.T) {
	t.Parallel()

	dthc := NewDataTrieHashesCollector()
	numHashes := 1000
	wg := &sync.WaitGroup{}
	wg.Add(numHashes)
	for i := 0; i < numHashes; i++ {
		go func(index int) {
			dthc.AddDirtyHash([]byte(strconv.Itoa(index)))
			wg.Done()
		}(i)
	}
	wg.Wait()

	for i := 0; i < numHashes; i++ {
		_, ok := dthc.newHashes[strconv.Itoa(i)]
		assert.True(t, ok)
	}
}

func TestDataTrieHashesCollector_GetDirtyHashes(t *testing.T) {
	t.Parallel()

	dthc := NewDataTrieHashesCollector()
	numHashes := 1000
	for i := 0; i < numHashes; i++ {
		dthc.AddDirtyHash([]byte(strconv.Itoa(i)))
	}

	dirtyHashes := dthc.GetDirtyHashes()
	assert.Equal(t, numHashes, len(dirtyHashes))
	for i := 0; i < numHashes; i++ {
		_, ok := dirtyHashes[strconv.Itoa(i)]
		assert.True(t, ok)
	}
}

func TestDataTrieHashesCollector_AddObsoleteHashes(t *testing.T) {
	t.Parallel()

	dthc := NewDataTrieHashesCollector()
	numHashes := 1000
	dirtyHashes := make([][]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		dirtyHashes[i] = []byte(strconv.Itoa(i))
	}

	dthc.AddObsoleteHashes(nil, dirtyHashes)

	assert.Equal(t, numHashes, len(dthc.oldHashes))
	for i := 0; i < numHashes; i++ {
		_, ok := dthc.oldHashes[strconv.Itoa(i)]
		assert.True(t, ok)
	}
}

func TestDataTriehashesCollector_GetCollectedData(t *testing.T) {
	t.Parallel()

	dthc := NewDataTrieHashesCollector()
	numHashes := 1000
	dirtyHashes := make([][]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		dirtyHashes[i] = []byte(strconv.Itoa(i))
		dthc.AddDirtyHash(dirtyHashes[i])
	}
	dthc.AddObsoleteHashes(nil, dirtyHashes)

	oldRootHash, oldHashes, newHashes := dthc.GetCollectedData()
	assert.Nil(t, oldRootHash)
	assert.Equal(t, numHashes, len(oldHashes))
	assert.Equal(t, numHashes, len(newHashes))
	for i := 0; i < numHashes; i++ {
		_, ok := oldHashes[strconv.Itoa(i)]
		assert.True(t, ok)
		_, ok = newHashes[strconv.Itoa(i)]
		assert.True(t, ok)
	}
}

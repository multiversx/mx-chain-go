package evictionWaitingList

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/stretchr/testify/assert"
)

func getDefaultArgsForMemoryEvictionWaitingList() MemoryEvictionWaitingListArgs {
	return MemoryEvictionWaitingListArgs{
		RootHashesSize: 10,
		HashesSize:     10,
	}
}

func TestNewMemoryEvictionWaitingList(t *testing.T) {
	t.Parallel()

	mewl, err := NewMemoryEvictionWaitingList(getDefaultArgsForMemoryEvictionWaitingList())
	assert.Nil(t, err)
	assert.False(t, check.IfNil(mewl))
}

func TestNewMemoryEvictionWaitingList_HashesSize(t *testing.T) {
	t.Parallel()

	args := getDefaultArgsForMemoryEvictionWaitingList()
	args.HashesSize = 0
	mewl, err := NewMemoryEvictionWaitingList(args)
	assert.True(t, check.IfNil(mewl))
	assert.True(t, errors.Is(err, data.ErrInvalidCacheSize))
	assert.True(t, strings.Contains(err.Error(), "HashesSize"))
}

func TestNewMemoryEvictionWaitingList_RootHashesSize(t *testing.T) {
	t.Parallel()

	args := getDefaultArgsForMemoryEvictionWaitingList()
	args.RootHashesSize = 0
	mewl, err := NewMemoryEvictionWaitingList(args)
	assert.True(t, check.IfNil(mewl))
	assert.True(t, errors.Is(err, data.ErrInvalidCacheSize))
	assert.True(t, strings.Contains(err.Error(), "RootHashesSize"))
}

func TestMemoryEvictionWaitingList_Put(t *testing.T) {
	t.Parallel()

	mewl, _ := NewMemoryEvictionWaitingList(getDefaultArgsForMemoryEvictionWaitingList())

	hashesMap := &rootHashData{
		numReferences: 1,
		hashes:        [][]byte{[]byte("hash1"), []byte("hash2")},
	}
	root := []byte("root")

	err := mewl.Put(root, hashesMap.hashes)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(mewl.cache))
	assert.Equal(t, hashesMap, mewl.cache[string(root)])
}

func TestMemoryEvictionWaitingList_PutMultiple(t *testing.T) {
	t.Parallel()

	args := getDefaultArgsForMemoryEvictionWaitingList()
	args.RootHashesSize = 2
	args.HashesSize = 10000
	mewl, _ := NewMemoryEvictionWaitingList(args)

	hashesMap := &rootHashData{
		numReferences: 1,
		hashes:        [][]byte{[]byte("hash0"), []byte("hash1")},
	}
	roots := [][]byte{
		[]byte("root0"),
		[]byte("root1"),
		[]byte("root2"),
		[]byte("root3"),
		[]byte("root4"),
		[]byte("root5"),
	}

	for i := range roots {
		err := mewl.Put(roots[i], hashesMap.hashes)
		assert.Nil(t, err)
	}

	assert.Equal(t, 0, len(mewl.cache)) // 2 resets
	_ = mewl.Put(roots[0], hashesMap.hashes)
	assert.Equal(t, hashesMap, mewl.cache[string(roots[0])])
}

func TestMemoryEvictionWaitingList_PutMultipleCleanDB(t *testing.T) {
	t.Parallel()

	args := getDefaultArgsForMemoryEvictionWaitingList()
	args.RootHashesSize = 10000
	args.HashesSize = 2
	mewl, _ := NewMemoryEvictionWaitingList(args)

	hashesMap := &rootHashData{
		numReferences: 1,
		hashes:        [][]byte{[]byte("hash0"), []byte("hash1"), []byte("hash2")},
	}
	roots := [][]byte{
		[]byte("root0"),
		[]byte("root1"),
		[]byte("root2"),
		[]byte("root3"),
	}

	for i := range roots {
		err := mewl.Put(roots[i], hashesMap.hashes)
		assert.Nil(t, err)
	}

	assert.Equal(t, 0, len(mewl.cache)) // 4 reset
}

func TestMemoryEvictionWaitingList_Evict(t *testing.T) {
	t.Parallel()

	mewl, _ := NewMemoryEvictionWaitingList(getDefaultArgsForMemoryEvictionWaitingList())

	expectedHashesMap := &rootHashData{
		numReferences: 1,
		hashes:        [][]byte{[]byte("hash1"), []byte("hash2")},
	}
	root1 := []byte("root1")

	_ = mewl.Put(root1, expectedHashesMap.hashes)

	evicted, err := mewl.Evict([]byte("root1"))
	assert.Nil(t, err)
	assert.Equal(t, 0, len(mewl.cache))
	assert.Equal(t, expectedHashesMap.hashes, evicted)
}

func TestMemoryEvictionWaitingList_EvictFromDB(t *testing.T) {
	t.Parallel()

	args := getDefaultArgsForMemoryEvictionWaitingList()
	args.RootHashesSize = 4
	mewl, _ := NewMemoryEvictionWaitingList(args)

	hashesMap := &rootHashData{
		numReferences: 1,
		hashes:        [][]byte{[]byte("hash0"), []byte("hash1")},
	}
	roots := [][]byte{
		[]byte("root0"),
		[]byte("root1"),
		[]byte("root2"),
	}

	for i := range roots {
		_ = mewl.Put(roots[i], hashesMap.hashes)
	}

	vals, err := mewl.Evict(roots[2])
	assert.Nil(t, err)
	assert.Equal(t, hashesMap.hashes, vals)
}

func TestMemoryEvictionWaitingList_ShouldKeepHash(t *testing.T) {
	t.Parallel()

	mewl, _ := NewMemoryEvictionWaitingList(getDefaultArgsForMemoryEvictionWaitingList())

	hashesMap := &rootHashData{
		numReferences: 1,
		hashes:        [][]byte{[]byte("hash0"), []byte("hash1")},
	}
	roots := [][]byte{
		{1, 2, 3, 4, 5, 0},
		{6, 7, 8, 9, 10, 0},
		{1, 2, 3, 4, 5, 1},
	}

	for i := range roots {
		_ = mewl.Put(roots[i], hashesMap.hashes)
	}

	present, err := mewl.ShouldKeepHash("hash0", 1)
	assert.True(t, present)
	assert.Nil(t, err)
}

func TestMemoryEvictionWaitingList_ShouldKeepHashShouldReturnFalse(t *testing.T) {
	t.Parallel()

	mewl, _ := NewMemoryEvictionWaitingList(getDefaultArgsForMemoryEvictionWaitingList())

	hashesMap := &rootHashData{
		numReferences: 1,
		hashes:        [][]byte{[]byte("hash0"), []byte("hash1")},
	}
	roots := [][]byte{
		{1, 2, 3, 4, 5, 0},
		{6, 7, 8, 9, 10, 0},
	}

	for i := range roots {
		_ = mewl.Put(roots[i], hashesMap.hashes)
	}

	present, err := mewl.ShouldKeepHash("hash2", 1)
	assert.False(t, present)
	assert.Nil(t, err)
}

func TestMemoryEvictionWaitingList_ShouldKeepHashShouldReturnTrueIfPresentInOldHashes(t *testing.T) {
	t.Parallel()

	mewl, _ := NewMemoryEvictionWaitingList(getDefaultArgsForMemoryEvictionWaitingList())

	hashesMap := &rootHashData{
		numReferences: 1,
		hashes:        [][]byte{[]byte("hash0"), []byte("hash1")},
	}
	roots := [][]byte{
		{1, 2, 3, 4, 5, 0},
		{6, 7, 8, 9, 10, 0},
	}

	for i := range roots {
		_ = mewl.Put(roots[i], hashesMap.hashes)
	}

	present, err := mewl.ShouldKeepHash("hash0", 0)
	assert.False(t, present)
	assert.Nil(t, err)
}

func TestMemoryEvictionWaitingList_ShouldKeepHashSearchInDb(t *testing.T) {
	t.Parallel()

	args := getDefaultArgsForMemoryEvictionWaitingList()
	args.RootHashesSize = 2
	args.HashesSize = 10000
	mewl, _ := NewMemoryEvictionWaitingList(args)

	root1 := []byte{1, 2, 3, 4, 5, 0}
	root2 := []byte{6, 7, 8, 9, 10, 0}
	root3 := []byte{1, 2, 3, 4, 5, 1}
	root4 := []byte{1, 2, 3, 4, 5, 1}

	hashesMapSlice := []*rootHashData{
		{
			numReferences: 1,
			hashes:        [][]byte{[]byte("hash2"), []byte("hash3")},
		},
		{
			numReferences: 1,
			hashes:        [][]byte{[]byte("hash4"), []byte("hash5")},
		},
		{
			numReferences: 1,
			hashes:        [][]byte{[]byte("hash0"), []byte("hash1")},
		},
		{
			numReferences: 1,
			hashes:        [][]byte{[]byte("hash-1"), []byte("hash-2")},
		},
	}

	roots := [][]byte{
		root1,
		root2,
		root3,
		root4,
	}

	for i := range roots {
		_ = mewl.Put(roots[i], hashesMapSlice[i].hashes)
	}

	present, err := mewl.ShouldKeepHash("hash-1", 1)
	assert.True(t, present)
	assert.Nil(t, err)
}

func TestMemoryEvictionWaitingList_ShouldKeepHashInvalidKey(t *testing.T) {
	t.Parallel()

	mewl, _ := NewMemoryEvictionWaitingList(getDefaultArgsForMemoryEvictionWaitingList())

	hashesMap := &rootHashData{
		numReferences: 1,
		hashes:        [][]byte{[]byte("hash0"), []byte("hash1")},
	}

	_ = mewl.Put([]byte{}, hashesMap.hashes)

	present, err := mewl.ShouldKeepHash("hash0", 1)
	assert.False(t, present)
	assert.Equal(t, state.ErrInvalidKey, err)
}

func TestMemoryEvictionWaitingList_Close(t *testing.T) {
	t.Parallel()

	mewl, _ := NewMemoryEvictionWaitingList(getDefaultArgsForMemoryEvictionWaitingList())
	err := mewl.Close()
	assert.Nil(t, err)
}

func TestMemoryEvictionWaitingList_RemoveFromInversedCache(t *testing.T) {
	t.Parallel()

	roothash1 := "roothash1"
	roothash2 := "roothash2"
	roothash3 := "roothash3"
	hash := "hash"

	mewl, _ := NewMemoryEvictionWaitingList(getDefaultArgsForMemoryEvictionWaitingList())
	mewl.reversedCache[hash] = &hashInfo{
		roothashes: [][]byte{[]byte(roothash1), []byte(roothash2), []byte(roothash3)},
	}

	mewl.removeFromReversedCache([]byte(roothash1), [][]byte{[]byte(hash)})
	info := mewl.reversedCache[hash]
	assert.Equal(t,
		&hashInfo{
			roothashes: [][]byte{[]byte(roothash2), []byte(roothash3)},
		},
		info)

	mewl.removeFromReversedCache([]byte(roothash3), [][]byte{[]byte(hash)})
	info = mewl.reversedCache[hash]
	assert.Equal(t,
		&hashInfo{
			roothashes: [][]byte{[]byte(roothash2)},
		},
		info)

	mewl.removeFromReversedCache([]byte(roothash3), [][]byte{[]byte(hash)})
	info = mewl.reversedCache[hash]
	assert.Equal(t,
		&hashInfo{
			roothashes: [][]byte{[]byte(roothash2)},
		},
		info)

	mewl.removeFromReversedCache([]byte(roothash2), [][]byte{[]byte(hash)})
	info, exists := mewl.reversedCache[hash]
	assert.Nil(t, info)
	assert.False(t, exists)
}

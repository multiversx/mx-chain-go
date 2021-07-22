package evictionWaitingList

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go-logger/check"
	"github.com/ElrondNetwork/elrond-go/mock"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	"github.com/stretchr/testify/assert"
)

func getDefaultParametersForMemoryEvictionWaitingList() (uint, marshal.Marshalizer) {
	return 10, &mock.MarshalizerMock{}
}

func TestNewEvictionWaitingListV2(t *testing.T) {
	t.Parallel()

	mewl, err := NewMemoryEvictionWaitingListV2(getDefaultParametersForMemoryEvictionWaitingList())
	assert.Nil(t, err)
	assert.False(t, check.IfNil(mewl))
}

func TestNewEvictionWaitingListV2_InvalidCacheSize(t *testing.T) {
	t.Parallel()

	_, marsh := getDefaultParametersForMemoryEvictionWaitingList()
	mewl, err := NewMemoryEvictionWaitingListV2(0, marsh)
	assert.True(t, check.IfNil(mewl))
	assert.Equal(t, data.ErrInvalidCacheSize, err)
}

func TestNewEvictionWaitingListV2_NilDMarshalizer(t *testing.T) {
	t.Parallel()

	size, _ := getDefaultParametersForMemoryEvictionWaitingList()
	mewl, err := NewMemoryEvictionWaitingListV2(size, nil)
	assert.True(t, check.IfNil(mewl))
	assert.Equal(t, data.ErrNilMarshalizer, err)
}

func TestEvictionWaitingListV2_Put(t *testing.T) {
	t.Parallel()

	mewl, _ := NewMemoryEvictionWaitingListV2(getDefaultParametersForMemoryEvictionWaitingList())

	hashesMap := temporary.ModifiedHashes{
		"hash1": {},
		"hash2": {},
	}
	root := []byte("root")

	err := mewl.Put(root, hashesMap)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(mewl.cache))
	assert.Equal(t, hashesMap, mewl.cache[string(root)])
}

func TestEvictionWaitingListV2_PutMultiple(t *testing.T) {
	t.Parallel()

	cacheSize := uint(2)
	_, marsh := getDefaultParametersForMemoryEvictionWaitingList()
	mewl, _ := NewMemoryEvictionWaitingListV2(cacheSize, marsh)

	hashesMap := temporary.ModifiedHashes{
		"hash0": {},
		"hash1": {},
	}
	roots := [][]byte{
		[]byte("root0"),
		[]byte("root1"),
		[]byte("root2"),
		[]byte("root3"),
	}

	for i := range roots {
		err := mewl.Put(roots[i], hashesMap)
		assert.Nil(t, err)
	}

	assert.Equal(t, 0, len(mewl.cache)) // 2 resets
	_ = mewl.Put(roots[0], hashesMap)
	assert.Equal(t, hashesMap, mewl.cache[string(roots[0])])
}

func TestEvictionWaitingListV2_Evict(t *testing.T) {
	t.Parallel()

	mewl, _ := NewMemoryEvictionWaitingListV2(getDefaultParametersForMemoryEvictionWaitingList())

	expectedHashesMap := temporary.ModifiedHashes{
		"hash1": {},
		"hash2": {},
	}
	root1 := []byte("root1")

	_ = mewl.Put(root1, expectedHashesMap)

	evicted, err := mewl.Evict([]byte("root1"))
	assert.Nil(t, err)
	assert.Equal(t, 0, len(mewl.cache))
	assert.Equal(t, expectedHashesMap, evicted)
}

func TestEvictionWaitingListV2_EvictFromDB(t *testing.T) {
	t.Parallel()

	cacheSize := uint(4)
	_, marsh := getDefaultParametersForMemoryEvictionWaitingList()
	mewl, _ := NewMemoryEvictionWaitingListV2(cacheSize, marsh)

	hashesMap := temporary.ModifiedHashes{
		"hash0": {},
		"hash1": {},
	}
	roots := [][]byte{
		[]byte("root0"),
		[]byte("root1"),
		[]byte("root2"),
	}

	for i := range roots {
		_ = mewl.Put(roots[i], hashesMap)
	}

	vals, err := mewl.Evict(roots[2])
	assert.Nil(t, err)
	assert.Equal(t, hashesMap, vals)
}

func TestEvictionWaitingListV2_ShouldKeepHash(t *testing.T) {
	t.Parallel()

	mewl, _ := NewMemoryEvictionWaitingListV2(getDefaultParametersForMemoryEvictionWaitingList())

	hashesMap := temporary.ModifiedHashes{
		"hash0": {},
		"hash1": {},
	}
	roots := [][]byte{
		{1, 2, 3, 4, 5, 0},
		{6, 7, 8, 9, 10, 0},
		{1, 2, 3, 4, 5, 1},
	}

	for i := range roots {
		_ = mewl.Put(roots[i], hashesMap)
	}

	present, err := mewl.ShouldKeepHash("hash0", 1)
	assert.True(t, present)
	assert.Nil(t, err)
}

func TestEvictionWaitingListV2_ShouldKeepHashShouldReturnFalse(t *testing.T) {
	t.Parallel()

	mewl, _ := NewMemoryEvictionWaitingListV2(getDefaultParametersForMemoryEvictionWaitingList())

	hashesMap := temporary.ModifiedHashes{
		"hash0": {},
		"hash1": {},
	}
	roots := [][]byte{
		{1, 2, 3, 4, 5, 0},
		{6, 7, 8, 9, 10, 0},
	}

	for i := range roots {
		_ = mewl.Put(roots[i], hashesMap)
	}

	present, err := mewl.ShouldKeepHash("hash2", 1)
	assert.False(t, present)
	assert.Nil(t, err)
}

func TestEvictionWaitingListV2_ShouldKeepHashShouldReturnTrueIfPresentInOldHashes(t *testing.T) {
	t.Parallel()

	mewl, _ := NewMemoryEvictionWaitingListV2(getDefaultParametersForMemoryEvictionWaitingList())

	hashesMap := temporary.ModifiedHashes{
		"hash0": {},
		"hash1": {},
	}
	roots := [][]byte{
		{1, 2, 3, 4, 5, 0},
		{6, 7, 8, 9, 10, 0},
	}

	for i := range roots {
		_ = mewl.Put(roots[i], hashesMap)
	}

	present, err := mewl.ShouldKeepHash("hash0", 0)
	assert.False(t, present)
	assert.Nil(t, err)
}

func TestEvictionWaitingListV2_ShouldKeepHashSearchInDb(t *testing.T) {
	t.Parallel()

	cacheSize := uint(2)
	_, marsh := getDefaultParametersForMemoryEvictionWaitingList()
	mewl, _ := NewMemoryEvictionWaitingListV2(cacheSize, marsh)

	root1 := []byte{1, 2, 3, 4, 5, 0}
	root2 := []byte{6, 7, 8, 9, 10, 0}
	root3 := []byte{1, 2, 3, 4, 5, 1}

	hashesMapSlice := []temporary.ModifiedHashes{
		{
			"hash2": {},
			"hash3": {},
		},
		{
			"hash4": {},
			"hash5": {},
		},
		{
			"hash0": {},
			"hash1": {},
		},
	}
	roots := [][]byte{
		root1,
		root2,
		root3,
	}

	for i := range roots {
		_ = mewl.Put(roots[i], hashesMapSlice[i])
	}

	present, err := mewl.ShouldKeepHash("hash0", 1)
	assert.True(t, present)
	assert.Nil(t, err)
}

func TestEvictionWaitingListV2_ShouldKeepHashInvalidKey(t *testing.T) {
	t.Parallel()

	mewl, _ := NewMemoryEvictionWaitingListV2(getDefaultParametersForMemoryEvictionWaitingList())

	hashesMap := temporary.ModifiedHashes{
		"hash0": {},
		"hash1": {},
	}

	_ = mewl.Put([]byte{}, hashesMap)

	present, err := mewl.ShouldKeepHash("hash0", 1)
	assert.False(t, present)
	assert.Equal(t, state.ErrInvalidKey, err)
}

func TestNewEvictionWaitingListV2_Close(t *testing.T) {
	t.Parallel()

	mewl, _ := NewMemoryEvictionWaitingListV2(getDefaultParametersForMemoryEvictionWaitingList())
	err := mewl.Close()
	assert.Nil(t, err)
}

func TestEvictionWaitingList_RemoveFromInversedCache(t *testing.T) {
	t.Parallel()

	roothash1 := "roothash1"
	roothash2 := "roothash2"
	roothash3 := "roothash3"
	hash := "hash"
	modified := temporary.ModifiedHashes{
		hash: struct{}{},
	}

	mewl, _ := NewMemoryEvictionWaitingListV2(getDefaultParametersForMemoryEvictionWaitingList())
	mewl.reversedCache[hash] = &hashInfo{
		roothashes: [][]byte{[]byte(roothash1), []byte(roothash2), []byte(roothash3)},
	}

	mewl.removeFromReversedCache([]byte(roothash1), modified)
	info := mewl.reversedCache[hash]
	assert.Equal(t,
		&hashInfo{
			roothashes: [][]byte{[]byte(roothash2), []byte(roothash3)},
		},
		info)

	mewl.removeFromReversedCache([]byte(roothash3), modified)
	info = mewl.reversedCache[hash]
	assert.Equal(t,
		&hashInfo{
			roothashes: [][]byte{[]byte(roothash2)},
		},
		info)

	mewl.removeFromReversedCache([]byte(roothash3), modified)
	info = mewl.reversedCache[hash]
	assert.Equal(t,
		&hashInfo{
			roothashes: [][]byte{[]byte(roothash2)},
		},
		info)

	mewl.removeFromReversedCache([]byte(roothash2), modified)
	info, exists := mewl.reversedCache[hash]
	assert.Nil(t, info)
	assert.False(t, exists)
}

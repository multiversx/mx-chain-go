package evictionWaitingList

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go/mock"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/stretchr/testify/assert"
)

func TestNewEvictionWaitingListV2(t *testing.T) {
	t.Parallel()

	ec, err := NewEvictionWaitingListV2(getDefaultParameters())
	assert.Nil(t, err)
	assert.NotNil(t, ec)
}

func TestNewEvictionWaitingListV2_InvalidCacheSize(t *testing.T) {
	t.Parallel()

	_, db, marsh := getDefaultParameters()
	ec, err := NewEvictionWaitingListV2(0, db, marsh)
	assert.Nil(t, ec)
	assert.Equal(t, data.ErrInvalidCacheSize, err)
}

func TestNewEvictionWaitingListV2_NilDatabase(t *testing.T) {
	t.Parallel()

	size, _, marsh := getDefaultParameters()
	ec, err := NewEvictionWaitingListV2(size, nil, marsh)
	assert.Nil(t, ec)
	assert.Equal(t, data.ErrNilDatabase, err)
}

func TestNewEvictionWaitingListV2_NilDMarshalizer(t *testing.T) {
	t.Parallel()

	size, db, _ := getDefaultParameters()
	ec, err := NewEvictionWaitingListV2(size, db, nil)
	assert.Nil(t, ec)
	assert.Equal(t, data.ErrNilMarshalizer, err)
}

func TestEvictionWaitingListV2_Put(t *testing.T) {
	t.Parallel()

	ec, _ := NewEvictionWaitingListV2(getDefaultParameters())

	hashesMap := temporary.ModifiedHashes{
		"hash1": {},
		"hash2": {},
	}
	root := []byte("root")

	err := ec.Put(root, hashesMap)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(ec.cache))
	assert.Equal(t, hashesMap, ec.cache[string(root)])
}

func TestEvictionWaitingListV2_PutMultiple(t *testing.T) {
	t.Parallel()

	cacheSize := uint(2)
	_, db, marsh := getDefaultParameters()
	ec, _ := NewEvictionWaitingListV2(cacheSize, db, marsh)

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
		err := ec.Put(roots[i], hashesMap)
		assert.Nil(t, err)
	}

	assert.Equal(t, 4, len(ec.cache))
	for i := uint(0); i < cacheSize; i++ {
		assert.Equal(t, hashesMap, ec.cache[string(roots[i])])
	}
	for i := cacheSize; i < uint(len(roots)); i++ {
		encVal, err := ec.db.Get(roots[i])
		assert.Nil(t, err)

		b := &batch.Batch{}
		err = ec.marshalizer.Unmarshal(b, encVal)
		assert.Nil(t, err)

		val := make(temporary.ModifiedHashes, len(b.Data))

		for _, h := range b.Data {
			val[string(h)] = struct{}{}
		}

		assert.Equal(t, hashesMap, val)
	}
}

func TestEvictionWaitingListV2_Evict(t *testing.T) {
	t.Parallel()

	ec, _ := NewEvictionWaitingListV2(getDefaultParameters())

	expectedHashesMap := temporary.ModifiedHashes{
		"hash1": {},
		"hash2": {},
	}
	root1 := []byte("root1")

	_ = ec.Put(root1, expectedHashesMap)

	evicted, err := ec.Evict([]byte("root1"))
	assert.Nil(t, err)
	assert.Equal(t, 0, len(ec.cache))
	assert.Equal(t, expectedHashesMap, evicted)
}

func TestEvictionWaitingListV2_EvictFromDB(t *testing.T) {
	t.Parallel()

	cacheSize := uint(2)
	_, db, marsh := getDefaultParameters()
	ec, _ := NewEvictionWaitingListV2(cacheSize, db, marsh)

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
		_ = ec.Put(roots[i], hashesMap)
	}

	val, _ := ec.db.Get(roots[2])
	assert.NotNil(t, val)

	vals, err := ec.Evict(roots[2])
	assert.Nil(t, err)
	assert.Equal(t, hashesMap, vals)

	val, _ = ec.db.Get(roots[2])
	assert.Nil(t, val)
}

func TestEvictionWaitingListV2_ShouldKeepHash(t *testing.T) {
	t.Parallel()

	ewl, _ := NewEvictionWaitingListV2(getDefaultParameters())

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
		_ = ewl.Put(roots[i], hashesMap)
	}

	present, err := ewl.ShouldKeepHash("hash0", 1)
	assert.True(t, present)
	assert.Nil(t, err)
}

func TestEvictionWaitingListV2_ShouldKeepHashShouldReturnFalse(t *testing.T) {
	t.Parallel()

	ewl, _ := NewEvictionWaitingListV2(getDefaultParameters())

	hashesMap := temporary.ModifiedHashes{
		"hash0": {},
		"hash1": {},
	}
	roots := [][]byte{
		{1, 2, 3, 4, 5, 0},
		{6, 7, 8, 9, 10, 0},
	}

	for i := range roots {
		_ = ewl.Put(roots[i], hashesMap)
	}

	present, err := ewl.ShouldKeepHash("hash2", 1)
	assert.False(t, present)
	assert.Nil(t, err)
}

func TestEvictionWaitingListV2_ShouldKeepHashShouldReturnTrueIfPresentInOldHashes(t *testing.T) {
	t.Parallel()

	ewl, _ := NewEvictionWaitingListV2(getDefaultParameters())

	hashesMap := temporary.ModifiedHashes{
		"hash0": {},
		"hash1": {},
	}
	roots := [][]byte{
		{1, 2, 3, 4, 5, 0},
		{6, 7, 8, 9, 10, 0},
	}

	for i := range roots {
		_ = ewl.Put(roots[i], hashesMap)
	}

	present, err := ewl.ShouldKeepHash("hash0", 0)
	assert.False(t, present)
	assert.Nil(t, err)
}

func TestEvictionWaitingListV2_ShouldKeepHashSearchInDb(t *testing.T) {
	t.Parallel()

	cacheSize := uint(2)
	_, db, marsh := getDefaultParameters()
	ewl, _ := NewEvictionWaitingListV2(cacheSize, db, marsh)

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
		_ = ewl.Put(roots[i], hashesMapSlice[i])
	}

	present, err := ewl.ShouldKeepHash("hash0", 1)
	assert.True(t, present)
	assert.Nil(t, err)
}

func TestEvictionWaitingListV2_ShouldKeepHashInvalidKey(t *testing.T) {
	t.Parallel()

	ewl, _ := NewEvictionWaitingListV2(getDefaultParameters())

	hashesMap := temporary.ModifiedHashes{
		"hash0": {},
		"hash1": {},
	}

	_ = ewl.Put([]byte{}, hashesMap)

	present, err := ewl.ShouldKeepHash("hash0", 1)
	assert.False(t, present)
	assert.Equal(t, state.ErrInvalidKey, err)
}

func TestNewEvictionWaitingListV2_Close(t *testing.T) {
	t.Parallel()

	db := memorydb.New()
	ewl, err := NewEvictionWaitingListV2(10, db, &mock.MarshalizerMock{})
	assert.Nil(t, err)
	assert.NotNil(t, ewl)

	err = ewl.Close()
	assert.Nil(t, err)
}

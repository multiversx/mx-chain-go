package evictionWaitingList

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func getDefaultParameters() (uint, storage.Persister, marshal.Marshalizer) {
	return 10, database.NewMemDB(), &testscommon.MarshalizerMock{}
}

func TestNewEvictionWaitingList(t *testing.T) {
	t.Parallel()

	ec, err := NewEvictionWaitingList(getDefaultParameters())
	assert.Nil(t, err)
	assert.NotNil(t, ec)
}

func TestNewEvictionWaitingList_InvalidCacheSize(t *testing.T) {
	t.Parallel()

	_, db, marsh := getDefaultParameters()
	ec, err := NewEvictionWaitingList(0, db, marsh)
	assert.Nil(t, ec)
	assert.Equal(t, data.ErrInvalidCacheSize, err)
}

func TestNewEvictionWaitingList_NilDatabase(t *testing.T) {
	t.Parallel()

	size, _, marsh := getDefaultParameters()
	ec, err := NewEvictionWaitingList(size, nil, marsh)
	assert.Nil(t, ec)
	assert.Equal(t, data.ErrNilDatabase, err)
}

func TestNewEvictionWaitingList_NilDMarshalizer(t *testing.T) {
	t.Parallel()

	size, db, _ := getDefaultParameters()
	ec, err := NewEvictionWaitingList(size, db, nil)
	assert.Nil(t, ec)
	assert.Equal(t, data.ErrNilMarshalizer, err)
}

func TestEvictionWaitingList_Put(t *testing.T) {
	t.Parallel()

	ec, _ := NewEvictionWaitingList(getDefaultParameters())

	hashesMap := common.ModifiedHashes{
		"hash1": {},
		"hash2": {},
	}
	root := []byte("root")

	err := ec.Put(root, hashesMap)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(ec.cache))
	assert.Equal(t, hashesMap, ec.cache[string(root)])
}

func TestEvictionWaitingList_PutMultiple(t *testing.T) {
	t.Parallel()

	cacheSize := uint(2)
	_, db, marsh := getDefaultParameters()
	ec, _ := NewEvictionWaitingList(cacheSize, db, marsh)

	hashesMap := common.ModifiedHashes{
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

		val := make(common.ModifiedHashes, len(b.Data))

		for _, h := range b.Data {
			val[string(h)] = struct{}{}
		}

		assert.Equal(t, hashesMap, val)
	}
}

func TestEvictionWaitingList_Evict(t *testing.T) {
	t.Parallel()

	ec, _ := NewEvictionWaitingList(getDefaultParameters())

	expectedHashesMap := common.ModifiedHashes{
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

func TestEvictionWaitingList_EvictFromDB(t *testing.T) {
	t.Parallel()

	cacheSize := uint(2)
	_, db, marsh := getDefaultParameters()
	ec, _ := NewEvictionWaitingList(cacheSize, db, marsh)

	hashesMap := common.ModifiedHashes{
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

func TestEvictionWaitingList_ShouldKeepHash(t *testing.T) {
	t.Parallel()

	ewl, _ := NewEvictionWaitingList(getDefaultParameters())

	hashesMap := common.ModifiedHashes{
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

func TestEvictionWaitingList_ShouldKeepHashShouldReturnFalse(t *testing.T) {
	t.Parallel()

	ewl, _ := NewEvictionWaitingList(getDefaultParameters())

	hashesMap := common.ModifiedHashes{
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

func TestEvictionWaitingList_ShouldKeepHashShouldReturnTrueIfPresentInOldHashes(t *testing.T) {
	t.Parallel()

	ewl, _ := NewEvictionWaitingList(getDefaultParameters())

	hashesMap := common.ModifiedHashes{
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

func TestEvictionWaitingList_ShouldKeepHashSearchInDb(t *testing.T) {
	t.Parallel()

	cacheSize := uint(2)
	_, db, marsh := getDefaultParameters()
	ewl, _ := NewEvictionWaitingList(cacheSize, db, marsh)

	root1 := []byte{1, 2, 3, 4, 5, 0}
	root2 := []byte{6, 7, 8, 9, 10, 0}
	root3 := []byte{1, 2, 3, 4, 5, 1}

	hashesMapSlice := []common.ModifiedHashes{
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

func TestEvictionWaitingList_ShouldKeepHashInvalidKey(t *testing.T) {
	t.Parallel()

	ewl, _ := NewEvictionWaitingList(getDefaultParameters())

	hashesMap := common.ModifiedHashes{
		"hash0": {},
		"hash1": {},
	}

	_ = ewl.Put([]byte{}, hashesMap)

	present, err := ewl.ShouldKeepHash("hash0", 1)
	assert.False(t, present)
	assert.Equal(t, state.ErrInvalidKey, err)
}

func TestNewEvictionWaitingList_Close(t *testing.T) {
	t.Parallel()

	memDB := database.NewMemDB()
	ewl, err := NewEvictionWaitingList(10, memDB, &testscommon.MarshalizerMock{})
	assert.Nil(t, err)
	assert.NotNil(t, ewl)

	err = ewl.Close()
	assert.Nil(t, err)
}

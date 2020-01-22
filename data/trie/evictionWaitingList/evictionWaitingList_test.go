package evictionWaitingList

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/stretchr/testify/assert"
)

func getDefaultParameters() (uint, storage.Persister, marshal.Marshalizer) {
	return 10, memorydb.New(), &mock.MarshalizerMock{}
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

	hashesMap := map[string]struct{}{
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

	hashesMap := map[string]struct{}{
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
		val := make(map[string]struct{}, 0)
		encVal, err := ec.db.Get(roots[i])
		assert.Nil(t, err)

		err = ec.marshalizer.Unmarshal(&val, encVal)

		assert.Nil(t, err)
		assert.Equal(t, hashesMap, val)
	}

}

func TestEvictionWaitingList_Evict(t *testing.T) {
	t.Parallel()

	ec, _ := NewEvictionWaitingList(getDefaultParameters())

	expectedHashesMap := map[string]struct{}{
		"hash1": {},
		"hash2": {},
	}
	root1 := []byte("root1")

	_ = ec.Put(root1, expectedHashesMap)

	hashes, err := ec.Evict([]byte("root1"))
	assert.Nil(t, err)
	assert.Equal(t, 0, len(ec.cache))
	assert.Equal(t, expectedHashesMap, hashes)
}

func TestEvictionWaitingList_EvictFromDB(t *testing.T) {
	t.Parallel()

	cacheSize := uint(2)
	_, db, marsh := getDefaultParameters()
	ec, _ := NewEvictionWaitingList(cacheSize, db, marsh)

	hashesMap := map[string]struct{}{
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

func TestEvictionWaitingList_PresentInNewHashes(t *testing.T) {
	t.Parallel()

	ewl, _ := NewEvictionWaitingList(getDefaultParameters())

	hashesMap := map[string]struct{}{
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

	present, err := ewl.PresentInNewHashes("hash0")
	assert.True(t, present)
	assert.Nil(t, err)
}

func TestEvictionWaitingList_PresentInNewHashesShouldReturnFalse(t *testing.T) {
	t.Parallel()

	ewl, _ := NewEvictionWaitingList(getDefaultParameters())

	hashesMap := map[string]struct{}{
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

	present, err := ewl.PresentInNewHashes("hash0")
	assert.False(t, present)
	assert.Nil(t, err)
}

func TestEvictionWaitingList_PresentInNewHashesInDb(t *testing.T) {
	t.Parallel()

	cacheSize := uint(2)
	_, db, marsh := getDefaultParameters()
	ewl, _ := NewEvictionWaitingList(cacheSize, db, marsh)

	root1 := []byte{1, 2, 3, 4, 5, 0}
	root2 := []byte{6, 7, 8, 9, 10, 0}
	root3 := []byte{1, 2, 3, 4, 5, 1}

	hashesMap := map[string]struct{}{
		"hash0": {},
		"hash1": {},
	}
	roots := [][]byte{
		root1,
		root2,
		root3,
	}

	for i := range roots {
		_ = ewl.Put(roots[i], hashesMap)
	}

	present, err := ewl.PresentInNewHashes("hash0")
	assert.True(t, present)
	assert.Nil(t, err)
}

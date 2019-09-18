package evictionWaitingList

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/stretchr/testify/assert"
)

func getDefaultParameters() (int, storage.Persister, marshal.Marshalizer) {
	return 10, memorydb.New(), &mock.MarshalizerMock{}
}

func TestNewEvictionWaitingList(t *testing.T) {
	ec, err := NewEvictionWaitingList(getDefaultParameters())
	assert.Nil(t, err)
	assert.NotNil(t, ec)
}

func TestNewEvictionWaitingList_InvalidCacheSize(t *testing.T) {
	_, db, marsh := getDefaultParameters()
	ec, err := NewEvictionWaitingList(0, db, marsh)
	assert.Nil(t, ec)
	assert.Equal(t, trie.ErrInvalidCacheSize, err)
}

func TestNewEvictionWaitingList_NilDatabase(t *testing.T) {
	size, _, marsh := getDefaultParameters()
	ec, err := NewEvictionWaitingList(size, nil, marsh)
	assert.Nil(t, ec)
	assert.Equal(t, trie.ErrNilDatabase, err)
}

func TestNewEvictionWaitingList_NilDMarshalizer(t *testing.T) {
	size, db, _ := getDefaultParameters()
	ec, err := NewEvictionWaitingList(size, db, nil)
	assert.Nil(t, ec)
	assert.Equal(t, trie.ErrNilMarshalizer, err)
}

func TestEvictionCache_Add(t *testing.T) {
	ec, _ := NewEvictionWaitingList(getDefaultParameters())

	hashes := [][]byte{
		[]byte("hash1"),
		[]byte("hash2"),
	}
	root := []byte("root")

	err := ec.Add(root, hashes)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(ec.cache))
	assert.Equal(t, hashes, ec.cache[string(root)])
}

func TestEvictionCache_AddMultiple(t *testing.T) {
	cacheSize := 2
	_, db, marsh := getDefaultParameters()
	ec, _ := NewEvictionWaitingList(cacheSize, db, marsh)

	hashes := [][]byte{
		[]byte("hash0"),
		[]byte("hash1"),
	}
	roots := [][]byte{
		[]byte("root0"),
		[]byte("root1"),
		[]byte("root2"),
		[]byte("root3"),
	}

	for i := range roots {
		err := ec.Add(roots[i], hashes)
		assert.Nil(t, err)
	}

	assert.Equal(t, 2, len(ec.cache))
	for i := 0; i < cacheSize; i++ {
		assert.Equal(t, hashes, ec.cache[string(roots[i])])
	}
	for i := cacheSize; i < len(roots); i++ {
		val := make([][]byte, 0)
		encVal, err := ec.db.Get(roots[i])
		err = ec.marshalizer.Unmarshal(&val, encVal)

		assert.Nil(t, err)
		assert.Equal(t, hashes, val)
	}

}

func TestEvictionCache_Evict(t *testing.T) {
	ec, _ := NewEvictionWaitingList(getDefaultParameters())

	expectedHashes := [][]byte{
		[]byte("hash1"),
		[]byte("hash2"),
	}
	root1 := []byte("root1")

	_ = ec.Add(root1, expectedHashes)

	hashes, err := ec.Evict([]byte("root1"))
	assert.Nil(t, err)
	assert.Equal(t, 0, len(ec.cache))
	assert.Equal(t, expectedHashes, hashes)
}

func TestEvictionCache_EvictFromDB(t *testing.T) {
	cacheSize := 2
	_, db, marsh := getDefaultParameters()
	ec, _ := NewEvictionWaitingList(cacheSize, db, marsh)

	hashes := [][]byte{
		[]byte("hash0"),
		[]byte("hash1"),
	}
	roots := [][]byte{
		[]byte("root0"),
		[]byte("root1"),
		[]byte("root2"),
	}

	for i := range roots {
		_ = ec.Add(roots[i], hashes)
	}

	val, _ := ec.db.Get(roots[2])
	assert.NotNil(t, val)

	vals, err := ec.Evict(roots[2])
	assert.Nil(t, err)
	assert.Equal(t, hashes, vals)

	val, _ = ec.db.Get(roots[2])
	assert.Nil(t, val)
}

func TestEvictionCache_Rollback(t *testing.T) {
	ec, _ := NewEvictionWaitingList(getDefaultParameters())

	h1 := [][]byte{
		[]byte("hash1"),
		[]byte("hash2"),
	}
	h2 := [][]byte{
		[]byte("hash3"),
		[]byte("hash4"),
	}
	root1 := []byte("root1")
	root2 := []byte("root2")

	_ = ec.Add(root1, h1)
	_ = ec.Add(root2, h2)

	err := ec.Rollback(root1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ec.cache))
	assert.Equal(t, h2, ec.cache[string(root2)])
}

func TestEvictionCache_RollbackFromDB(t *testing.T) {
	cacheSize := 2
	_, db, marsh := getDefaultParameters()
	ec, _ := NewEvictionWaitingList(cacheSize, db, marsh)

	hashes := [][]byte{
		[]byte("hash0"),
		[]byte("hash1"),
	}
	roots := [][]byte{
		[]byte("root0"),
		[]byte("root1"),
		[]byte("root2"),
	}

	for i := range roots {
		_ = ec.Add(roots[i], hashes)
	}

	val, _ := ec.db.Get(roots[2])
	assert.NotNil(t, val)

	err := ec.Rollback(roots[2])
	assert.Nil(t, err)

	val, _ = ec.db.Get(roots[2])
	assert.Nil(t, val)
}

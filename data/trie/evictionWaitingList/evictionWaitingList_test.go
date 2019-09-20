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

func getDefaultParameters() (int, storage.Persister, marshal.Marshalizer) {
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

	hashes := [][]byte{
		[]byte("hash1"),
		[]byte("hash2"),
	}
	root := []byte("root")

	err := ec.Put(root, hashes)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(ec.cache))
	assert.Equal(t, hashes, ec.cache[string(root)])
}

func TestEvictionWaitingList_PutMultiple(t *testing.T) {
	t.Parallel()

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
		err := ec.Put(roots[i], hashes)
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

func TestEvictionWaitingList_Evict(t *testing.T) {
	t.Parallel()

	ec, _ := NewEvictionWaitingList(getDefaultParameters())

	expectedHashes := [][]byte{
		[]byte("hash1"),
		[]byte("hash2"),
	}
	root1 := []byte("root1")

	_ = ec.Put(root1, expectedHashes)

	hashes, err := ec.Evict([]byte("root1"))
	assert.Nil(t, err)
	assert.Equal(t, 0, len(ec.cache))
	assert.Equal(t, expectedHashes, hashes)
}

func TestEvictionWaitingList_EvictFromDB(t *testing.T) {
	t.Parallel()

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
		_ = ec.Put(roots[i], hashes)
	}

	val, _ := ec.db.Get(roots[2])
	assert.NotNil(t, val)

	vals, err := ec.Evict(roots[2])
	assert.Nil(t, err)
	assert.Equal(t, hashes, vals)

	val, _ = ec.db.Get(roots[2])
	assert.Nil(t, val)
}

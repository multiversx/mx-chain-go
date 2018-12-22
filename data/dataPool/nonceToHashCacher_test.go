package dataPool_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/stretchr/testify/assert"
)

//------- NewNonceToHashCacher

func TestNewNonceToHashCacher_NilCacherShouldErr(t *testing.T) {
	t.Parallel()

	_, err := dataPool.NewNonceToHashCacher(nil, mock.NewNonceHashConverterMock())
	assert.Equal(t, data.ErrNilCacher, err)
}

func TestNewNonceToHashCacher_NilConverterShouldErr(t *testing.T) {
	t.Parallel()

	cacher, _ := storage.NewCache(storage.LRUCache, 10000)

	_, err := dataPool.NewNonceToHashCacher(cacher, nil)
	assert.Equal(t, data.ErrNilNonceConverter, err)
}

func TestNewNonceToHashCacher_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	cacher, _ := storage.NewCache(storage.LRUCache, 10000)

	_, err := dataPool.NewNonceToHashCacher(cacher, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)
}

//------- test called methods

func TestNonceToHashCacher_ClearShouldBeCalled(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.LRUCacheStub{}

	wasCalled := false

	mockLRU.ClearCalled = func() {
		wasCalled = true
	}

	nthc, err := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)

	nthc.Clear()
	assert.True(t, wasCalled)
}

func TestNonceToHashCacher_PutShouldBeCalled(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.LRUCacheStub{}

	wasCalled := false

	mockLRU.PutCalled = func(key []byte, value interface{}) (evicted bool) {
		wasCalled = true

		return true
	}

	nthc, err := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)

	nthc.Put(6, []byte("aaaa"))
	assert.True(t, wasCalled)
}

func TestNonceToHashCacher_HasShouldBeCalled(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.LRUCacheStub{}

	wasCalled := false

	mockLRU.HasCalled = func(key []byte) bool {
		wasCalled = true

		return true
	}

	nthc, err := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)

	nthc.Has(6)
	assert.True(t, wasCalled)
}

func TestNonceToHashCacher_HasOrAddShouldBeCalled(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.LRUCacheStub{}

	wasCalled := false

	mockLRU.HasOrAddCalled = func(key []byte, value interface{}) (bool, bool) {
		wasCalled = true

		return true, true
	}

	nthc, err := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)

	nthc.HasOrAdd(6, nil)
	assert.True(t, wasCalled)
}

func TestNonceToHashCacher_RemoveShouldBeCalled(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.LRUCacheStub{}

	wasCalled := false

	mockLRU.RemoveCalled = func(key []byte) {
		wasCalled = true
	}

	nthc, err := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)

	nthc.Remove(6)
	assert.True(t, wasCalled)
}

func TestNonceToHashCacher_RemoveOldestShouldBeCalled(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.LRUCacheStub{}

	wasCalled := false

	mockLRU.RemoveOldestCalled = func() {
		wasCalled = true
	}

	nthc, err := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)

	nthc.RemoveOldest()
	assert.True(t, wasCalled)
}

func TestNonceToHashCacher_LenShouldBeCalled(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.LRUCacheStub{}

	wasCalled := false

	mockLRU.LenCalled = func() int {
		wasCalled = true

		return 0
	}

	nthc, err := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)

	assert.Equal(t, 0, nthc.Len())
	assert.True(t, wasCalled)
}

//------- Get

func TestNonceToHashCacher_GetNotFoundShouldRetNil(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.LRUCacheStub{}

	wasCalled := false

	mockLRU.GetCalled = func(key []byte) (value interface{}, ok bool) {
		wasCalled = true
		return nil, false
	}

	nthc, err := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)

	hash, _ := nthc.Get(6)
	assert.Nil(t, hash)
	assert.True(t, wasCalled)
}

func TestNonceToHashCacher_GetFoundShouldRetValue(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.LRUCacheStub{}

	wasCalled := false

	mockLRU.GetCalled = func(key []byte) (value interface{}, ok bool) {
		wasCalled = true
		return []byte("bbbb"), false
	}

	nthc, err := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)

	hash, _ := nthc.Get(6)
	assert.Equal(t, []byte("bbbb"), hash)
	assert.True(t, wasCalled)
}

//------- Peek

func TestNonceToHashCacher_PeekNotFoundShouldRetNil(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.LRUCacheStub{}

	wasCalled := false

	mockLRU.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		wasCalled = true
		return nil, false
	}

	nthc, err := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)

	hash, _ := nthc.Peek(6)
	assert.Nil(t, hash)
	assert.True(t, wasCalled)
}

func TestNonceToHashCacher_PeekFoundShouldRetValue(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.LRUCacheStub{}

	wasCalled := false

	mockLRU.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		wasCalled = true
		return []byte("bbbb"), false
	}

	nthc, err := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)

	hash, _ := nthc.Peek(6)
	assert.Equal(t, []byte("bbbb"), hash)
	assert.True(t, wasCalled)
}

//------- Keys

func TestNonceToHashCacher_KeysEmptyShouldRetEmpty(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.LRUCacheStub{}

	wasCalled := false

	mockLRU.KeysCalled = func() [][]byte {
		wasCalled = true
		return make([][]byte, 0)
	}

	nthc, err := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)

	keys := nthc.Keys()
	assert.NotNil(t, keys)
	assert.True(t, wasCalled)
	assert.Equal(t, 0, len(keys))
}

func TestNonceToHashCacher_KeysFoundShouldRetValues(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.LRUCacheStub{}

	wasCalled := false

	converter := mock.NewNonceHashConverterMock()
	nthc, err := dataPool.NewNonceToHashCacher(mockLRU, converter)
	assert.Nil(t, err)

	mockLRU.KeysCalled = func() [][]byte {
		wasCalled = true

		keys := make([][]byte, 0)
		keys = append(keys, converter.ToByteSlice(6))
		keys = append(keys, converter.ToByteSlice(10))

		return keys
	}

	nonces := nthc.Keys()
	assert.Equal(t, 2, len(nonces))
	assert.True(t, wasCalled)
	assert.Equal(t, uint64(6), nonces[0])
	assert.Equal(t, uint64(10), nonces[1])
}

//------- RegisterAddedDataHandler

func TestNonceToHashCacher_RegisterAddedDataHandlerShouldWork(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.LRUCacheStub{}

	wasCalled := false

	converter := mock.NewNonceHashConverterMock()
	nthc, err := dataPool.NewNonceToHashCacher(mockLRU, converter)
	assert.Nil(t, err)

	var handlerCacher func(key []byte)
	mockLRU.RegisterHandlerCalled = func(handler func(key []byte)) {
		handlerCacher = handler
	}

	nthc.RegisterHandler(func(nonce uint64) {
		wasCalled = true
		assert.Equal(t, uint64(6), nonce)
	})

	handlerCacher(converter.ToByteSlice(6))
	assert.True(t, wasCalled)
}

package dataPool_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/stretchr/testify/assert"
)

//------- NewNonceToHashCacher

func TestNewNonceToHashCacher_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	_, err := dataPool.NewNonceToHashCacher(storage.CacheConfig{
		Size: 1000,
		Type: storage.LRUCache,
	})
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

	nthc := dataPool.NewTestNonceToHashCacher(mockLRU)

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

	nthc := dataPool.NewTestNonceToHashCacher(mockLRU)

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

	nthc := dataPool.NewTestNonceToHashCacher(mockLRU)

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

	nthc := dataPool.NewTestNonceToHashCacher(mockLRU)

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

	nthc := dataPool.NewTestNonceToHashCacher(mockLRU)

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

	nthc := dataPool.NewTestNonceToHashCacher(mockLRU)

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

	nthc := dataPool.NewTestNonceToHashCacher(mockLRU)

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

	nthc := dataPool.NewTestNonceToHashCacher(mockLRU)

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

	nthc := dataPool.NewTestNonceToHashCacher(mockLRU)

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

	nthc := dataPool.NewTestNonceToHashCacher(mockLRU)

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

	nthc := dataPool.NewTestNonceToHashCacher(mockLRU)

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

	nthc := dataPool.NewTestNonceToHashCacher(mockLRU)

	keys := nthc.Keys()
	assert.NotNil(t, keys)
	assert.True(t, wasCalled)
	assert.Equal(t, 0, len(keys))
}

func TestNonceToHashCacher_KeysFoundShouldRetValues(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.LRUCacheStub{}

	wasCalled := false

	nthc := dataPool.NewTestNonceToHashCacher(mockLRU)

	mockLRU.KeysCalled = func() [][]byte {
		wasCalled = true

		keys := make([][]byte, 0)
		keys = append(keys, nthc.NonceToByteArray(6))
		keys = append(keys, nthc.NonceToByteArray(10))

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

	nthc := dataPool.NewTestNonceToHashCacher(mockLRU)

	var handlerCacher func(key []byte)
	mockLRU.RegisterHandlerCalled = func(handler func(key []byte)) {
		handlerCacher = handler
	}

	nthc.RegisterHandler(func(nonce uint64) {
		wasCalled = true
		assert.Equal(t, uint64(6), nonce)
	})

	handlerCacher(nthc.NonceToByteArray(6))
	assert.True(t, wasCalled)
}

func TestNonceToHashCacher_Uint64ToByteArrayConversion(t *testing.T) {
	t.Parallel()

	nthc := dataPool.NonceToHashCacher{}

	vals := make(map[uint64][]byte)

	vals[uint64(0)] = []byte{0, 0, 0, 0, 0, 0, 0, 0}
	vals[uint64(1)] = []byte{0, 0, 0, 0, 0, 0, 0, 1}
	vals[uint64(255)] = []byte{0, 0, 0, 0, 0, 0, 0, 255}
	vals[uint64(256)] = []byte{0, 0, 0, 0, 0, 0, 1, 0}
	vals[uint64(65536)] = []byte{0, 0, 0, 0, 0, 1, 0, 0}
	vals[uint64(1<<64-1)] = []byte{255, 255, 255, 255, 255, 255, 255, 255}

	for k, v := range vals {
		buff := nthc.NonceToByteArray(k)

		assert.Equal(t, v, buff)

		fmt.Printf("%v converts to: %v, got %v\n", k, v, buff)
	}
}

func BenchmarkNonceToHashCacher_Uint64ToByteArrayConversion(b *testing.B) {
	nthc := dataPool.NonceToHashCacher{}

	for i := 0; i < b.N; i++ {
		_ = nthc.NonceToByteArray(uint64(i))
	}
}

func BenchmarkNonceToHashCacher_Uint64ToByteArrayConversionAndBackToUint64(b *testing.B) {
	nthc := dataPool.NonceToHashCacher{}

	for i := 0; i < b.N; i++ {
		buff := nthc.NonceToByteArray(uint64(i))
		val := nthc.ByteArrayToNonce(buff)

		if uint64(i) != val {
			assert.Fail(b, fmt.Sprintf("Not equal %v, got %v\n", i, val))
		}
	}
}

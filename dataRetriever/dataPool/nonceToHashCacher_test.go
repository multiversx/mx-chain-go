package dataPool_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/stretchr/testify/assert"
)

//------- NewNonceToHashCacher

func TestNewNonceToHashCacher_NilCacherShouldErr(t *testing.T) {
	t.Parallel()

	nthc, err := dataPool.NewNonceToHashCacher(nil, mock.NewNonceHashConverterMock())
	assert.Equal(t, data.ErrNilCacher, err)
	assert.Nil(t, nthc)
}

func TestNewNonceToHashCacher_NilConverterShouldErr(t *testing.T) {
	t.Parallel()

	nthc, err := dataPool.NewNonceToHashCacher(&mock.CacherStub{}, nil)
	assert.Equal(t, data.ErrNilNonceConverter, err)
	assert.Nil(t, nthc)
}

func TestNewNonceToHashCacher_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	nthc, err := dataPool.NewNonceToHashCacher(&mock.CacherStub{}, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)
	assert.NotNil(t, nthc)
}

//------- test called methods

func TestNonceToHashCacher_ClearShouldBeCalled(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.CacherStub{}

	wasCalled := false

	mockLRU.ClearCalled = func() {
		wasCalled = true
	}

	nthc, _ := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())

	nthc.Clear()
	assert.True(t, wasCalled)
}

func TestNonceToHashCacher_PutShouldBeCalled(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.CacherStub{}

	wasCalled := false

	mockLRU.PutCalled = func(key []byte, value interface{}) (evicted bool) {
		wasCalled = true

		return true
	}

	nthc, _ := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())

	assert.True(t, nthc.Put(6, []byte("aaaa")))
	assert.True(t, wasCalled)
}

func TestNonceToHashCacher_HasShouldBeCalled(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.CacherStub{}

	wasCalled := false

	mockLRU.HasCalled = func(key []byte) bool {
		wasCalled = true

		return true
	}

	nthc, _ := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())

	assert.True(t, nthc.Has(6))
	assert.True(t, wasCalled)
}

func TestNonceToHashCacher_HasOrAddShouldBeCalled(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.CacherStub{}

	wasCalled := false

	mockLRU.HasOrAddCalled = func(key []byte, value interface{}) (bool, bool) {
		wasCalled = true

		return true, true
	}

	nthc, _ := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())

	ok, evicted := nthc.HasOrAdd(6, nil)

	assert.True(t, ok)
	assert.True(t, evicted)
	assert.True(t, wasCalled)
}

func TestNonceToHashCacher_RemoveShouldBeCalled(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.CacherStub{}

	wasCalled := false

	mockLRU.RemoveCalled = func(key []byte) {
		wasCalled = true
	}

	nthc, _ := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())

	nthc.Remove(6)
	assert.True(t, wasCalled)
}

func TestNonceToHashCacher_RemoveOldestShouldBeCalled(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.CacherStub{}

	wasCalled := false

	mockLRU.RemoveOldestCalled = func() {
		wasCalled = true
	}

	nthc, _ := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())

	nthc.RemoveOldest()
	assert.True(t, wasCalled)
}

func TestNonceToHashCacher_LenShouldBeCalled(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.CacherStub{}

	wasCalled := false

	mockLRU.LenCalled = func() int {
		wasCalled = true

		return 0
	}

	nthc, _ := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())

	assert.Equal(t, 0, nthc.Len())
	assert.True(t, wasCalled)
}

//------- Get

func TestNonceToHashCacher_GetNotFoundShouldRetNil(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.CacherStub{}

	wasCalled := false

	mockLRU.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		wasCalled = true
		return nil, false
	}

	nthc, _ := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())

	hash, ok := nthc.Get(6)
	assert.Nil(t, hash)
	assert.False(t, ok)
	assert.True(t, wasCalled)
}

func TestNonceToHashCacher_GetFoundShouldRetValue(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.CacherStub{}

	wasCalled := false

	mockLRU.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		wasCalled = true
		return []byte("bbbb"), true
	}

	nthc, _ := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())

	hash, ok := nthc.Get(6)
	assert.True(t, ok)
	assert.Equal(t, []byte("bbbb"), hash)
	assert.True(t, wasCalled)
}

//------- Peek

func TestNonceToHashCacher_PeekNotFoundShouldRetNil(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.CacherStub{}

	wasCalled := false

	mockLRU.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		wasCalled = true
		return nil, false
	}

	nthc, _ := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())

	hash, ok := nthc.Peek(6)
	assert.False(t, ok)
	assert.Nil(t, hash)
	assert.True(t, wasCalled)
}

func TestNonceToHashCacher_PeekFoundShouldRetValue(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.CacherStub{}

	wasCalled := false

	mockLRU.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		wasCalled = true
		return []byte("bbbb"), true
	}

	nthc, _ := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())

	hash, _ := nthc.Peek(6)
	assert.Equal(t, []byte("bbbb"), hash)
	assert.True(t, wasCalled)
}

//------- Keys

func TestNonceToHashCacher_KeysEmptyShouldRetEmpty(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.CacherStub{}

	mockLRU.KeysCalled = func() [][]byte {
		return make([][]byte, 0)
	}

	nthc, _ := dataPool.NewNonceToHashCacher(mockLRU, mock.NewNonceHashConverterMock())

	keys := nthc.Keys()
	assert.NotNil(t, keys)
	assert.Equal(t, 0, len(keys))
}

func TestNonceToHashCacher_KeysFoundShouldRetValues(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.CacherStub{}

	converter := mock.NewNonceHashConverterMock()
	nthc, _ := dataPool.NewNonceToHashCacher(mockLRU, converter)

	mockLRU.KeysCalled = func() [][]byte {
		keys := make([][]byte, 0)
		keys = append(keys, converter.ToByteSlice(6))
		keys = append(keys, converter.ToByteSlice(10))

		return keys
	}

	nonces := nthc.Keys()
	assert.Equal(t, 2, len(nonces))
	assert.Equal(t, uint64(6), nonces[0])
	assert.Equal(t, uint64(10), nonces[1])
}

func TestNonceToHashCacher_KeysWasCalled(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.CacherStub{}

	wasCalled := false

	converter := mock.NewNonceHashConverterMock()
	nthc, _ := dataPool.NewNonceToHashCacher(mockLRU, converter)

	mockLRU.KeysCalled = func() [][]byte {
		wasCalled = true

		return nil
	}

	assert.Equal(t, make([]uint64, 0), nthc.Keys())
	assert.True(t, wasCalled)
}

//------- RegisterAddedDataHandler

func TestNonceToHashCacher_RegisterAddedDataHandlerWithNilShouldNotRegister(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.CacherStub{}

	wasRegistered := false

	converter := mock.NewNonceHashConverterMock()
	nthc, _ := dataPool.NewNonceToHashCacher(mockLRU, converter)

	mockLRU.RegisterHandlerCalled = func(handler func(key []byte)) {
		wasRegistered = true
	}

	nthc.RegisterHandler(nil)

	assert.False(t, wasRegistered)
}

func TestNonceToHashCacher_RegisterAddedDataHandlerWithHandlerShouldRegister(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.CacherStub{}

	wasRegistered := false

	converter := mock.NewNonceHashConverterMock()
	nthc, _ := dataPool.NewNonceToHashCacher(mockLRU, converter)

	mockLRU.RegisterHandlerCalled = func(handler func(key []byte)) {
		wasRegistered = true
	}

	nthc.RegisterHandler(func(nonce uint64) {

	})

	assert.True(t, wasRegistered)
}

func TestNonceToHashCacher_RegisterAddedDataHandlerWithHandlerShouldBeCalled(t *testing.T) {
	t.Parallel()

	mockLRU := &mock.CacherStub{}

	wasCalled := false

	converter := mock.NewNonceHashConverterMock()
	nthc, _ := dataPool.NewNonceToHashCacher(mockLRU, converter)

	//save the handler so it can be called
	var h func(key []byte)

	mockLRU.RegisterHandlerCalled = func(handler func(key []byte)) {
		h = handler
	}

	nthc.RegisterHandler(func(nonce uint64) {
		if nonce == 67 {
			wasCalled = true
		}
	})

	h(converter.ToByteSlice(67))

	assert.True(t, wasCalled)
}

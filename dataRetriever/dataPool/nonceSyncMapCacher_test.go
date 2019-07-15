package dataPool_test

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/stretchr/testify/assert"
)

var timeToWaitAddedCallback = time.Second

func TestNewNonceSyncMapCacher_NilCacherShouldErr(t *testing.T) {
	t.Parallel()

	nsmc, err := dataPool.NewNonceSyncMapCacher(nil, &mock.Uint64ByteSliceConverterMock{})

	assert.Nil(t, nsmc)
	assert.Equal(t, dataRetriever.ErrNilCacher, err)
}

func TestNewNonceSyncMapCacher_SliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	nsmc, err := dataPool.NewNonceSyncMapCacher(&mock.CacherStub{}, nil)

	assert.Nil(t, nsmc)
	assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
}

func TestNewNonceSyncMapCacher_ShouldWork(t *testing.T) {
	t.Parallel()

	nsmc, err := dataPool.NewNonceSyncMapCacher(&mock.CacherStub{}, &mock.Uint64ByteSliceConverterMock{})

	assert.NotNil(t, nsmc)
	assert.Nil(t, err)
}

//------- Clear

func TestNonceSyncMapCacher_Clear(t *testing.T) {
	t.Parallel()

	wasCalled := false
	nsmc, _ := dataPool.NewNonceSyncMapCacher(
		&mock.CacherStub{
			ClearCalled: func() {
				wasCalled = true
			},
		},
		&mock.Uint64ByteSliceConverterMock{})

	nsmc.Clear()

	assert.True(t, wasCalled)
}

//------- Get

func TestNonceSyncMapGet_GetWhenMissingShouldReturnFalseNil(t *testing.T) {
	t.Parallel()

	nsmc, _ := dataPool.NewNonceSyncMapCacher(
		&mock.CacherStub{
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		},
		&mock.Uint64ByteSliceConverterMock{
			ToByteSliceCalled: func(u uint64) []byte {
				return make([]byte, 0)
			},
		})

	syncMap, ok := nsmc.Get(0)

	assert.False(t, ok)
	assert.Nil(t, syncMap)
}

func TestNonceSyncMapGet_GetWithWrongObjectTypeShouldReturnFalseNil(t *testing.T) {
	t.Parallel()

	nsmc, _ := dataPool.NewNonceSyncMapCacher(
		&mock.CacherStub{
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				return "wrong object type", true
			},
		},
		&mock.Uint64ByteSliceConverterMock{
			ToByteSliceCalled: func(u uint64) []byte {
				return make([]byte, 0)
			},
		})

	syncMap, ok := nsmc.Get(0)

	assert.False(t, ok)
	assert.Nil(t, syncMap)
}

func TestNonceSyncMapGet_GetShouldReturnTruePointer(t *testing.T) {
	t.Parallel()

	existingMap := &dataPool.ShardIdHashSyncMap{}
	nsmc, _ := dataPool.NewNonceSyncMapCacher(
		&mock.CacherStub{
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				return existingMap, true
			},
		},
		&mock.Uint64ByteSliceConverterMock{
			ToByteSliceCalled: func(u uint64) []byte {
				return make([]byte, 0)
			},
		})

	syncMap, ok := nsmc.Get(0)

	assert.True(t, ok)
	assert.True(t, syncMap == existingMap)
}

//------- RegisterHandler

func TestNonceSyncMapGet_RegisterHandlerWithNilPointerShouldNotAddHandler(t *testing.T) {
	t.Parallel()

	nsmc, _ := dataPool.NewNonceSyncMapCacher(&mock.CacherStub{}, &mock.Uint64ByteSliceConverterMock{})

	nsmc.RegisterHandler(nil)
	handlers := nsmc.GetAddedHandlers()

	assert.Empty(t, handlers)
}

func TestNonceSyncMapGet_RegisterHandlerShouldAddHandler(t *testing.T) {
	t.Parallel()

	nsmc, _ := dataPool.NewNonceSyncMapCacher(&mock.CacherStub{}, &mock.Uint64ByteSliceConverterMock{})

	handler := func(nonce uint64, shardId uint32, value []byte) {}

	nsmc.RegisterHandler(handler)
	handlers := nsmc.GetAddedHandlers()

	assert.Equal(t, 1, len(handlers))
	assert.NotNil(t, handlers[0])
}

//------- RemoveNonce

func TestNonceSyncMapCacher_RemoveNonce(t *testing.T) {
	t.Parallel()

	wasCalled := false
	nsmc, _ := dataPool.NewNonceSyncMapCacher(
		&mock.CacherStub{
			RemoveCalled: func(key []byte) {
				wasCalled = true
			},
		},
		mock.NewNonceHashConverterMock())

	nonce := uint64(40)
	nsmc.RemoveNonce(nonce)

	assert.True(t, wasCalled)
}

//------- RemoveShardId

func TestNonceSyncMapCacher_RemoveShardIdWhenSyncMapDoesNotExistsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not fail: %v", r))
		}
	}()

	shardId := uint32(89)
	nonce := uint64(45)

	nsmc, _ := dataPool.NewNonceSyncMapCacher(
		&mock.CacherStub{
			PeekCalled: func(key []byte) (interface{}, bool) {
				return nil, false
			},
		},
		mock.NewNonceHashConverterMock())

	nsmc.RemoveShardId(nonce, shardId)
}

func TestNonceSyncMapCacher_RemoveShardIdWhenExists(t *testing.T) {
	t.Parallel()

	shardId := uint32(89)
	nonce := uint64(45)
	syncMap := &dataPool.ShardIdHashSyncMap{}
	syncMap.Store(shardId, []byte("hash"))

	nsmc, _ := dataPool.NewNonceSyncMapCacher(
		&mock.CacherStub{
			PeekCalled: func(key []byte) (interface{}, bool) {
				return syncMap, true
			},
		},
		mock.NewNonceHashConverterMock())

	nsmc.RemoveShardId(nonce, shardId)

	syncMap.Range(func(shardId uint32, hash []byte) bool {
		assert.Fail(t, "should have not found the existing key")
		return false
	})
}

//------- Merge

func TestNonceSyncMapCacher_MergeNilSrcShouldIgnore(t *testing.T) {
	t.Parallel()

	cacher := mock.NewCacherMock()
	nsmc, _ := dataPool.NewNonceSyncMapCacher(cacher, &mock.Uint64ByteSliceConverterMock{})

	addedWasCalled := make(chan struct{})
	handler := func(nonce uint64, shardId uint32, value []byte) {
		addedWasCalled <- struct{}{}
	}

	nsmc.RegisterHandler(handler)
	nonce := uint64(40)
	nsmc.Merge(nonce, nil)

	assert.Equal(t, 0, cacher.Len())

	select {
	case <-addedWasCalled:
		assert.Fail(t, "should have not called added")
	case <-time.After(timeToWaitAddedCallback):
	}
}

func TestNonceSyncMapCacher_MergeWithEmptyMapShouldCreatesAnEmptyMapAndNotCallsAdded(t *testing.T) {
	t.Parallel()

	cacher := mock.NewCacherMock()
	nsmc, _ := dataPool.NewNonceSyncMapCacher(cacher, mock.NewNonceHashConverterMock())

	addedWasCalled := make(chan struct{})
	handler := func(nonce uint64, shardId uint32, value []byte) {
		addedWasCalled <- struct{}{}
	}

	nonce := uint64(40)
	nsmc.RegisterHandler(handler)
	nsmc.Merge(nonce, &dataPool.ShardIdHashSyncMap{})

	assert.Equal(t, 1, cacher.Len())

	retrievedMap, ok := nsmc.Get(nonce)

	assert.True(t, ok)
	numExpectedValsInSyncMap := 0
	testRetrievedMapAndAddedNotCalled(t, retrievedMap, addedWasCalled, numExpectedValsInSyncMap)
}

func TestNonceSyncMapCacher_MergeWithExistingKeyAndValShouldNotMergeOrCallAdded(t *testing.T) {
	t.Parallel()

	cacher := mock.NewCacherMock()
	nonceConverter := mock.NewNonceHashConverterMock()
	nsmc, _ := dataPool.NewNonceSyncMapCacher(cacher, nonceConverter)

	addedWasCalled := make(chan struct{})
	handler := func(nonce uint64, shardId uint32, value []byte) {
		addedWasCalled <- struct{}{}
	}

	nonce := uint64(40)
	shardId := uint32(7)
	val := []byte("value")
	existingMap := &dataPool.ShardIdHashSyncMap{}
	existingMap.Store(shardId, val)
	nonceHash := nonceConverter.ToByteSlice(nonce)
	cacher.Put(nonceHash, existingMap)

	nsmc.RegisterHandler(handler)
	syncMap := &dataPool.ShardIdHashSyncMap{}
	syncMap.Store(shardId, val)
	nsmc.Merge(nonce, syncMap)

	assert.Equal(t, 1, cacher.Len())

	retrievedMap, ok := nsmc.Get(nonce)

	assert.True(t, ok)
	numExpectedValsInSyncMap := 1
	testRetrievedMapAndAddedNotCalled(t, retrievedMap, addedWasCalled, numExpectedValsInSyncMap)
}

func TestNonceSyncMapCacher_MergeWithNewKeyShouldMergeAndCallAdded(t *testing.T) {
	t.Parallel()

	cacher := mock.NewCacherMock()
	nonceConverter := mock.NewNonceHashConverterMock()
	nsmc, _ := dataPool.NewNonceSyncMapCacher(cacher, nonceConverter)

	nonce := uint64(40)
	shardId := uint32(7)
	val := []byte("value")

	addedWasCalled := make(chan struct{})
	handler := func(addedNonce uint64, addedShardId uint32, addedValue []byte) {
		if addedNonce != nonce {
			assert.Fail(t, "invalid nonce retrieved")
			return
		}
		if addedShardId != shardId {
			assert.Fail(t, "invalid shardID retrieved")
			return
		}
		if !bytes.Equal(addedValue, val) {
			assert.Fail(t, "invalid value retrieved")
			return
		}

		addedWasCalled <- struct{}{}
	}

	nsmc.RegisterHandler(handler)
	syncMap := &dataPool.ShardIdHashSyncMap{}
	syncMap.Store(shardId, val)
	nsmc.Merge(nonce, syncMap)

	assert.Equal(t, 1, cacher.Len())

	retrievedMap, ok := nsmc.Get(nonce)

	assert.True(t, ok)
	numExpectedValsInSyncMap := 1
	testRetrievedMapAndAddedCalled(t, retrievedMap, addedWasCalled, numExpectedValsInSyncMap)
}

func TestNonceSyncMapCacher_MergeWithExistingKeyButDifferentValShouldOverwriteAndCallAdded(t *testing.T) {
	t.Parallel()

	cacher := mock.NewCacherMock()
	nonceConverter := mock.NewNonceHashConverterMock()
	nsmc, _ := dataPool.NewNonceSyncMapCacher(cacher, nonceConverter)

	nonce := uint64(40)
	shardId := uint32(7)
	val := []byte("value")
	existingMap := &dataPool.ShardIdHashSyncMap{}
	existingMap.Store(shardId, val)
	nonceHash := nonceConverter.ToByteSlice(nonce)
	cacher.Put(nonceHash, existingMap)

	newVal := []byte("new value")

	addedWasCalled := make(chan struct{})
	handler := func(addedNonce uint64, addedShardId uint32, addedValue []byte) {
		if addedNonce != nonce {
			assert.Fail(t, "invalid nonce retrieved")
			return
		}
		if addedShardId != shardId {
			assert.Fail(t, "invalid shardID retrieved")
			return
		}
		if !bytes.Equal(addedValue, newVal) {
			assert.Fail(t, "invalid value retrieved")
			return
		}

		addedWasCalled <- struct{}{}
	}

	nsmc.RegisterHandler(handler)
	syncMap := &dataPool.ShardIdHashSyncMap{}
	syncMap.Store(shardId, newVal)
	nsmc.Merge(nonce, syncMap)

	assert.Equal(t, 1, cacher.Len())

	retrievedMap, ok := nsmc.Get(nonce)

	assert.True(t, ok)
	numExpectedValsInSyncMap := 1
	testRetrievedMapAndAddedCalled(t, retrievedMap, addedWasCalled, numExpectedValsInSyncMap)
}

func TestNonceSyncMapCacher_MergeInConcurrentialSettingShouldWork(t *testing.T) {
	t.Parallel()

	cacher := mock.NewCacherMock()
	nonceConverter := mock.NewNonceHashConverterMock()
	nsmc, _ := dataPool.NewNonceSyncMapCacher(cacher, nonceConverter)

	wg := sync.WaitGroup{}
	nsmc.RegisterHandler(func(nonce uint64, shardId uint32, value []byte) {
		wg.Done()
	})

	maxNonces := 100
	maxShards := 25
	wg.Add(maxNonces * maxShards)

	chDone := make(chan struct{})
	go func() {
		wg.Wait()
		chDone <- struct{}{}
	}()

	for nonce := uint64(0); nonce < uint64(maxNonces); nonce++ {
		for shardId := uint32(0); shardId < uint32(maxShards); shardId++ {
			syncMap := &dataPool.ShardIdHashSyncMap{}
			syncMap.Store(shardId, make([]byte, 0))

			go nsmc.Merge(nonce, syncMap)
		}
	}

	select {
	case <-chDone:
	case <-time.After(timeToWaitAddedCallback * 10):
		assert.Fail(t, "timeout receiving all callbacks")
		return
	}

	for nonce := uint64(0); nonce < uint64(maxNonces); nonce++ {
		retrievedMap, ok := nsmc.Get(nonce)

		assert.True(t, ok)
		valsInsideMap := 0
		retrievedMap.Range(func(shardId uint32, hash []byte) bool {
			valsInsideMap++
			return true
		})

		assert.Equal(t, maxShards, valsInsideMap)
	}
}

//------- Has

func TestNonceSyncMapCacher_HasNotFoundShouldRetFalse(t *testing.T) {
	t.Parallel()

	cacher := mock.NewCacherMock()
	nonceConverter := mock.NewNonceHashConverterMock()
	nsmc, _ := dataPool.NewNonceSyncMapCacher(cacher, nonceConverter)

	inexistentNonce := uint64(67)
	has := nsmc.Has(inexistentNonce)

	assert.False(t, has)
}

func TestNonceSyncMapCacher_HasFoundShouldRetTrue(t *testing.T) {
	t.Parallel()

	cacher := mock.NewCacherMock()
	nonceConverter := mock.NewNonceHashConverterMock()
	nsmc, _ := dataPool.NewNonceSyncMapCacher(cacher, nonceConverter)

	nonce := uint64(67)
	nsmc.Merge(nonce, &dataPool.ShardIdHashSyncMap{})
	has := nsmc.Has(nonce)

	assert.True(t, has)
}

func testRetrievedMapAndAddedNotCalled(
	t *testing.T,
	retrievedMap dataRetriever.ShardIdHashMap,
	addedWasCalled chan struct{},
	numExpectedValsInSyncMap int,
) {

	foundVals := 0
	retrievedMap.Range(func(shardId uint32, hash []byte) bool {
		foundVals++
		return true
	})

	assert.Equal(t, numExpectedValsInSyncMap, foundVals)

	select {
	case <-addedWasCalled:
		assert.Fail(t, "should have not called added")
	case <-time.After(timeToWaitAddedCallback):
	}
}

func testRetrievedMapAndAddedCalled(
	t *testing.T,
	retrievedMap dataRetriever.ShardIdHashMap,
	addedWasCalled chan struct{},
	numExpectedValsInSyncMap int,
) {

	foundVals := 0
	retrievedMap.Range(func(shardId uint32, hash []byte) bool {
		foundVals++
		return true
	})

	assert.Equal(t, numExpectedValsInSyncMap, foundVals)

	select {
	case <-addedWasCalled:
	case <-time.After(timeToWaitAddedCallback):
		assert.Fail(t, "should have called added")
	}
}

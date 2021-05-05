package storageCacherAdapter

import (
	"fmt"
	"math"
	"testing"

	dataMock "github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

func TestStorageCacherAdapter_Clear(t *testing.T) {
	t.Parallel()

	clearCacheCalled := false
	su := &mock.StorerStub{
		ClearCacheCalled: func() {
			clearCacheCalled = true
		},
	}

	sca := NewStorageCacherAdapter(su)

	sca.Clear()
	assert.True(t, clearCacheCalled)
}

func TestStorageCacherAdapter_PutInvalidValueShouldNotPutInStorage(t *testing.T) {
	t.Parallel()

	putCalled := false
	su := &mock.StorerStub{
		PutCalled: func(_, _ []byte) error {
			putCalled = true
			return nil
		},
	}

	sca := NewStorageCacherAdapter(su)
	sca.Put([]byte("key"), 100, 100)

	assert.False(t, putCalled)
}

func TestStorageCacherAdapter_Put(t *testing.T) {
	t.Parallel()

	putCalled := false
	su := &mock.StorerStub{
		PutCalled: func(_, _ []byte) error {
			putCalled = true
			return nil
		},
	}

	sca := NewStorageCacherAdapter(su)
	sca.Put([]byte("key"), []byte("value"), 100)

	assert.True(t, putCalled)
}

func TestStorageCacherAdapter_GetErrShouldNotRetrieveVal(t *testing.T) {
	t.Parallel()

	value := []byte("value")
	returnedErr := fmt.Errorf("get error")
	su := &mock.StorerStub{
		GetCalled: func(_ []byte) ([]byte, error) {
			return value, returnedErr
		},
	}

	sca := NewStorageCacherAdapter(su)
	retrievedVal, _ := sca.Get([]byte("key"))

	assert.Nil(t, retrievedVal)
}

func TestStorageCacherAdapter_Get(t *testing.T) {
	t.Parallel()

	value := []byte("value")
	su := &mock.StorerStub{
		GetCalled: func(_ []byte) ([]byte, error) {
			return value, nil
		},
	}

	sca := NewStorageCacherAdapter(su)
	retrievedVal, _ := sca.Get([]byte("key"))

	retrievedValBytes, ok := retrievedVal.([]byte)
	assert.True(t, ok)
	assert.Equal(t, value, retrievedValBytes)
}

func TestStorageCacherAdapter_HasErrShouldReturnFalse(t *testing.T) {
	t.Parallel()

	returnedErr := fmt.Errorf("get error")
	su := &mock.StorerStub{
		HasCalled: func(_ []byte) error {
			return returnedErr
		},
	}

	sca := NewStorageCacherAdapter(su)
	isPresent := sca.Has([]byte("key"))

	assert.False(t, isPresent)
}

func TestStorageCacherAdapter_Has(t *testing.T) {
	t.Parallel()

	su := &mock.StorerStub{
		HasCalled: func(_ []byte) error {
			return nil
		},
	}

	sca := NewStorageCacherAdapter(su)
	isPresent := sca.Has([]byte("key"))

	assert.True(t, isPresent)
}

func TestStorageCacherAdapter_PeekCallsGet(t *testing.T) {
	t.Parallel()

	getCalled := false
	su := &mock.StorerStub{
		GetCalled: func(_ []byte) ([]byte, error) {
			getCalled = true
			return nil, nil
		},
	}

	sca := NewStorageCacherAdapter(su)
	_, _ = sca.Peek([]byte("key"))

	assert.True(t, getCalled)
}

func TestStorageCacherAdapter_HasOrAddCallsPut(t *testing.T) {
	t.Parallel()

	putCalled := false
	su := &mock.StorerStub{
		PutCalled: func(_, _ []byte) error {
			putCalled = true
			return nil
		},
	}

	sca := NewStorageCacherAdapter(su)
	_, _ = sca.HasOrAdd([]byte("key"), []byte("value"), 100)

	assert.True(t, putCalled)
}

func TestStorageCacherAdapter_Remove(t *testing.T) {
	t.Parallel()

	removeCalled := false
	su := &mock.StorerStub{
		RemoveCalled: func(_ []byte) error {
			removeCalled = true
			return nil
		},
	}

	sca := NewStorageCacherAdapter(su)
	sca.Remove([]byte("key"))

	assert.True(t, removeCalled)
}

func TestStorageCacherAdapter_Keys(t *testing.T) {
	t.Parallel()

	cacher, _ := lrucache.NewCache(2)
	db := dataMock.NewMemDbMock()
	su, _ := storageUnit.NewStorageUnit(cacher, db)

	sca := NewStorageCacherAdapter(su)

	_ = su.Put([]byte("key1"), []byte("val1"))
	_ = su.Put([]byte("key2"), []byte("val2"))
	_ = su.Put([]byte("key3"), []byte("val3"))

	keys := sca.Keys()
	assert.Equal(t, 3, len(keys))
}

func TestStorageCacherAdapter_Len(t *testing.T) {
	t.Parallel()

	cacher, _ := lrucache.NewCache(2)
	db := dataMock.NewMemDbMock()
	su, _ := storageUnit.NewStorageUnit(cacher, db)

	sca := NewStorageCacherAdapter(su)

	_ = su.Put([]byte("key1"), []byte("val1"))
	_ = su.Put([]byte("key2"), []byte("val2"))
	_ = su.Put([]byte("key3"), []byte("val3"))

	numElements := sca.Len()
	assert.Equal(t, 3, numElements)
}

func TestStorageCacherAdapter_SizeInBytesContained(t *testing.T) {
	t.Parallel()

	cacher, _ := lrucache.NewCache(2)
	db := dataMock.NewMemDbMock()
	su, _ := storageUnit.NewStorageUnit(cacher, db)

	sca := NewStorageCacherAdapter(su)

	_ = su.Put([]byte("key1"), []byte("val1"))
	_ = su.Put([]byte("key2"), []byte("val2"))
	_ = su.Put([]byte("key3"), []byte("val3"))

	totalElementsSize := sca.SizeInBytesContained()
	assert.Equal(t, uint64(12), totalElementsSize)
}

func TestStorageCacherAdapter_MaxSize(t *testing.T) {
	t.Parallel()

	sca := NewStorageCacherAdapter(&mock.StorerStub{})
	maxSize := sca.MaxSize()

	assert.Equal(t, math.MaxInt32, maxSize)
}

func TestStorageCacherAdapter_RegisterHandler(t *testing.T) {
	t.Parallel()

	sca := NewStorageCacherAdapter(&mock.StorerStub{})
	sca.RegisterHandler(nil, "")
}

func TestStorageCacherAdapter_UnRegisterHandler(t *testing.T) {
	t.Parallel()

	sca := NewStorageCacherAdapter(&mock.StorerStub{})
	sca.UnRegisterHandler("")
}

func TestStorageCacherAdapter_Close(t *testing.T) {
	t.Parallel()

	closeCalled := false
	su := &mock.StorerStub{
		CloseCalled: func() error {
			closeCalled = true
			return nil
		},
	}

	sca := NewStorageCacherAdapter(su)
	err := sca.Close()

	assert.Nil(t, err)
	assert.True(t, closeCalled)
}

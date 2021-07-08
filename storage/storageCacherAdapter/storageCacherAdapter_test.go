package storageCacherAdapter

import (
	"fmt"
	"math"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	dataMock "github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageMock "github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/ElrondNetwork/elrond-go/storage/storageCacherAdapter/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewStorageCacherAdapter_NilCacher(t *testing.T) {
	t.Parallel()

	sca, err := NewStorageCacherAdapter(
		nil,
		&storageMock.PersisterStub{},
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, sca)
	assert.Equal(t, storage.ErrNilCacher, err)
}

func TestNewStorageCacherAdapter_NilDB(t *testing.T) {
	t.Parallel()

	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{},
		nil,
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.True(t, check.IfNil(sca))
	assert.Equal(t, storage.ErrNilPersister, err)
}

func TestNewStorageCacherAdapter_NilStoredDataFactory(t *testing.T) {
	t.Parallel()

	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{},
		&storageMock.PersisterStub{},
		nil,
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, sca)
	assert.Equal(t, storage.ErrNilStoredDataFactory, err)
}

func TestNewStorageCacherAdapter_NilMarshalizer(t *testing.T) {
	t.Parallel()

	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{},
		&storageMock.PersisterStub{},
		factory.NewTrieNodeFactory(),
		nil,
	)
	assert.Nil(t, sca)
	assert.Equal(t, storage.ErrNilMarshalizer, err)
}

func TestStorageCacherAdapter_Clear(t *testing.T) {
	t.Parallel()

	purgeCalled := false
	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{
			PurgeCalled: func() {
				purgeCalled = true
			},
		},
		&storageMock.PersisterStub{},
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)

	sca.Clear()
	assert.True(t, purgeCalled)
}

func TestStorageCacherAdapter_Put(t *testing.T) {
	t.Parallel()

	addedKey := "key1"
	addedVal := []byte("value1")
	addSizedAndReturnEvictedCalled := false
	putCalled := false
	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{
			AddSizedAndReturnEvictedCalled: func(key, value interface{}, _ int64) map[interface{}]interface{} {
				stringKey, ok := key.(string)
				assert.True(t, ok)
				assert.Equal(t, addedKey, stringKey)

				res := make(map[interface{}]interface{})
				res[100] = 10
				res[stringKey] = value

				addSizedAndReturnEvictedCalled = true
				return res
			},
		},
		&storageMock.PersisterStub{
			PutCalled: func(key, _ []byte) error {
				assert.Equal(t, []byte(addedKey), key)
				putCalled = true
				return nil
			},
		},
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)

	evicted := sca.Put([]byte(addedKey), addedVal, 100)
	assert.True(t, evicted)
	assert.True(t, putCalled)
	assert.True(t, addSizedAndReturnEvictedCalled)
}

func TestStorageCacherAdapter_GetFoundInCacherShouldNotCallDbGet(t *testing.T) {
	t.Parallel()

	keyString := "key"
	cacherGetCalled := false
	dbGetCalled := false
	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{
			GetCalled: func(key interface{}) (interface{}, bool) {
				k, ok := key.(string)
				assert.True(t, ok)
				assert.Equal(t, keyString, k)

				cacherGetCalled = true
				return []byte("val"), true
			},
		},
		&storageMock.PersisterStub{
			GetCalled: func(_ []byte) ([]byte, error) {
				dbGetCalled = true
				return nil, nil
			},
		},
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)

	retrievedVal, _ := sca.Get([]byte(keyString))

	assert.Equal(t, []byte("val"), retrievedVal)
	assert.True(t, cacherGetCalled)
	assert.False(t, dbGetCalled)
}

type testStoredDataImpl struct {
}

type testStoredData struct {
	Key   []byte
	Value uint64
}

func (t *testStoredDataImpl) CreateEmpty() interface{} {
	return &testStoredData{}
}

func (t *testStoredDataImpl) IsInterfaceNil() bool {
	return t == nil
}

func TestStorageCacherAdapter_GetFromDb(t *testing.T) {
	t.Parallel()

	testData := testStoredData{
		Key:   []byte("key"),
		Value: 100,
	}

	marshalizer := &mock.MarshalizerMock{}
	cacherGetCalled := false
	dbGetCalled := false
	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{
			GetCalled: func(_ interface{}) (interface{}, bool) {
				cacherGetCalled = true
				return nil, false
			},
		},
		&storageMock.PersisterStub{
			GetCalled: func(_ []byte) ([]byte, error) {
				dbGetCalled = true
				byteData, err := marshalizer.Marshal(testData)
				return byteData, err
			},
		},
		&testStoredDataImpl{},
		marshalizer,
	)
	assert.Nil(t, err)

	retrievedVal, _ := sca.Get([]byte("key"))

	val, ok := retrievedVal.(*testStoredData)
	assert.True(t, ok)
	assert.Equal(t, testData.Key, val.Key)
	assert.Equal(t, testData.Value, val.Value)
	assert.True(t, cacherGetCalled)
	assert.True(t, dbGetCalled)
}

func TestStorageCacherAdapter_HasReturnsIfFoundInCacher(t *testing.T) {
	t.Parallel()

	containsCalled := false
	hasCalled := false
	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{
			ContainsCalled: func(key interface{}) bool {
				_, ok := key.(string)
				assert.True(t, ok)

				containsCalled = true
				return true
			},
		},
		&storageMock.PersisterStub{
			HasCalled: func(_ []byte) error {
				hasCalled = true
				return nil
			},
		},
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)

	isPresent := sca.Has([]byte("key"))

	assert.True(t, isPresent)
	assert.True(t, containsCalled)
	assert.False(t, hasCalled)
}

func TestStorageCacherAdapter_HasReturnsTrueIfFoundInDB(t *testing.T) {
	t.Parallel()

	containsCalled := false
	hasCalled := false
	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{
			ContainsCalled: func(_ interface{}) bool {
				containsCalled = true
				return false
			},
		},
		&storageMock.PersisterStub{
			HasCalled: func(_ []byte) error {
				hasCalled = true
				return nil
			},
		},
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)

	isPresent := sca.Has([]byte("key"))

	assert.True(t, isPresent)
	assert.True(t, containsCalled)
	assert.True(t, hasCalled)
}

func TestStorageCacherAdapter_HasReturnsFalseIfNotFound(t *testing.T) {
	t.Parallel()

	containsCalled := false
	hasCalled := false
	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{
			ContainsCalled: func(_ interface{}) bool {
				containsCalled = true
				return false
			},
		},
		&storageMock.PersisterStub{
			HasCalled: func(_ []byte) error {
				hasCalled = true
				return fmt.Errorf("not found err")
			},
		},
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)

	isPresent := sca.Has([]byte("key"))

	assert.False(t, isPresent)
	assert.True(t, containsCalled)
	assert.True(t, hasCalled)
}

func TestStorageCacherAdapter_Peek(t *testing.T) {
	t.Parallel()

	peekCalled := false
	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{
			PeekCalled: func(key interface{}) (interface{}, bool) {
				_, ok := key.(string)
				assert.True(t, ok)

				peekCalled = true
				return "value", true
			},
		},
		&storageMock.PersisterStub{},
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)

	val, ok := sca.Peek([]byte("key"))

	assert.True(t, peekCalled)
	assert.True(t, ok)
	assert.Equal(t, "value", val)
}

func TestStorageCacherAdapter_RemoveFromCacherFirst(t *testing.T) {
	t.Parallel()

	cacherRemoveCalled := false
	dbRemoveCalled := false
	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{
			RemoveCalled: func(key interface{}) bool {
				_, ok := key.(string)
				assert.True(t, ok)

				cacherRemoveCalled = true
				return true
			},
		},
		&storageMock.PersisterStub{
			RemoveCalled: func(key []byte) error {
				dbRemoveCalled = true
				return nil
			},
		},
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)

	sca.Remove([]byte("key"))

	assert.True(t, cacherRemoveCalled)
	assert.False(t, dbRemoveCalled)
}

func TestStorageCacherAdapter_RemoveFromDb(t *testing.T) {
	t.Parallel()

	cacherRemoveCalled := false
	dbRemoveCalled := false
	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{
			RemoveCalled: func(_ interface{}) bool {
				cacherRemoveCalled = true
				return false
			},
		},
		&storageMock.PersisterStub{
			RemoveCalled: func(_ []byte) error {
				dbRemoveCalled = true
				return nil
			},
		},
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)

	sca.Remove([]byte("key"))

	assert.True(t, cacherRemoveCalled)
	assert.True(t, dbRemoveCalled)
}

func TestStorageCacherAdapter_Keys(t *testing.T) {
	t.Parallel()

	db := dataMock.NewMemDbMock()
	_ = db.Put([]byte("key"), []byte("val"))
	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{
			KeysCalled: func() []interface{} {
				res := make([]interface{}, 0)
				res = append(res, "key2")
				return res
			},
		},
		db,
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)

	keys := sca.Keys()
	assert.Equal(t, 2, len(keys))
}

func TestStorageCacherAdapter_Len(t *testing.T) {
	t.Parallel()

	db := dataMock.NewMemDbMock()
	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{
			LenCalled: func() int {
				return 3
			},
			AddSizedAndReturnEvictedCalled: func(key, value interface{}, sizeInBytes int64) map[interface{}]interface{} {
				res := make(map[interface{}]interface{})
				res[key] = value
				return res
			},
		},
		db,
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)

	_ = sca.Put([]byte("key"), []byte("val"), 3)
	numVals := sca.Len()
	assert.Equal(t, 4, numVals)
}

func TestStorageCacherAdapter_SizeInBytesContained(t *testing.T) {
	t.Parallel()

	db := dataMock.NewMemDbMock()
	_ = db.Put([]byte("key"), []byte("val"))
	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{
			SizeInBytesContainedCalled: func() uint64 {
				return 1000
			},
		},
		db,
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)

	totalSize := sca.SizeInBytesContained()
	assert.Equal(t, uint64(1000), totalSize)
}

func TestStorageCacherAdapter_MaxSize(t *testing.T) {
	t.Parallel()

	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{},
		&storageMock.PersisterStub{},
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)

	maxSize := sca.MaxSize()
	assert.Equal(t, math.MaxInt64, maxSize)
}

func TestStorageCacherAdapter_RegisterHandler(t *testing.T) {
	t.Parallel()

	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{},
		&storageMock.PersisterStub{},
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)
	sca.RegisterHandler(nil, "")
}

func TestStorageCacherAdapter_UnRegisterHandler(t *testing.T) {
	t.Parallel()

	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{},
		&storageMock.PersisterStub{},
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)
	sca.UnRegisterHandler("")
}

func TestStorageCacherAdapter_Close(t *testing.T) {
	t.Parallel()

	closeCalled := false
	sca, err := NewStorageCacherAdapter(
		&storageMock.AdaptedSizedLruCacheStub{},
		&storageMock.PersisterStub{
			CloseCalled: func() error {
				closeCalled = true
				return nil
			},
		},
		factory.NewTrieNodeFactory(),
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)

	_ = sca.Close()
	assert.True(t, closeCalled)
}

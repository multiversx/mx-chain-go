package dataRetriever_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHas_ErrorWhenStorerIsMissing(t *testing.T) {
	t.Parallel()

	s := &storageStubs.StorerStub{}

	b := dataRetriever.NewChainStorer()
	assert.False(t, b.IsInterfaceNil())

	b.AddStorer(1, s)
	err := b.Has(2, []byte("whatever"))
	assert.Equal(t, dataRetriever.ErrNoSuchStorageUnit, err)
}

func TestHas_ReturnsCorrectly(t *testing.T) {
	t.Parallel()

	s := &storageStubs.StorerStub{}
	s.HasCalled = func(key []byte) error {
		return nil
	}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(1, s)
	err := b.Has(1, []byte("whatever"))
	assert.Nil(t, err)
}

func TestGet_ErrorWhenStorerIsMissing(t *testing.T) {
	t.Parallel()

	s := &storageStubs.StorerStub{}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(1, s)
	obj, err := b.Get(2, []byte("whatever"))
	assert.Nil(t, obj)
	assert.Equal(t, dataRetriever.ErrNoSuchStorageUnit, err)
}

func TestGet_ReturnsCorrectly(t *testing.T) {
	t.Parallel()

	s := &storageStubs.StorerStub{}
	getResult := []byte("get called")
	s.GetCalled = func(key []byte) (b []byte, e error) {
		return getResult, nil
	}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(1, s)
	obj, err := b.Get(1, []byte("whatever"))
	assert.Nil(t, err)
	assert.Equal(t, getResult, obj)
}

func TestPut_ErrorWhenStorerIsMissing(t *testing.T) {
	t.Parallel()

	s := &storageStubs.StorerStub{}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(1, s)
	err := b.Put(2, []byte("whatever"), []byte("whatever value"))

	assert.True(t, errors.Is(err, dataRetriever.ErrNoSuchStorageUnit))
}

func TestPut_ReturnsCorrectly(t *testing.T) {
	t.Parallel()

	s := &storageStubs.StorerStub{}
	putErr := errors.New("error")
	s.PutCalled = func(key, data []byte) error {
		return putErr
	}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(1, s)
	err := b.Put(1, []byte("whatever"), []byte("whatever value"))
	assert.Equal(t, putErr, err)
}

func TestGetAll_ErrorWhenStorerIsMissing(t *testing.T) {
	t.Parallel()

	s := &storageStubs.StorerStub{}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(1, s)
	ret, err := b.GetAll(2, [][]byte{[]byte("whatever"), []byte("whatever 2")})

	assert.Nil(t, ret)
	assert.Equal(t, dataRetriever.ErrNoSuchStorageUnit, err)
}

func TestGetAll_ErrorWhenStorersGetErrors(t *testing.T) {
	t.Parallel()

	s := &storageStubs.StorerStub{}
	getErr := errors.New("error")
	s.GetCalled = func(key []byte) (bytes []byte, e error) {
		return nil, getErr
	}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(1, s)
	ret, err := b.GetAll(1, [][]byte{[]byte("whatever"), []byte("whatever 2")})

	assert.Nil(t, ret)
	assert.Equal(t, getErr, err)
}

func TestGetAll_ReturnsCorrectly(t *testing.T) {
	t.Parallel()

	key1 := "key1"
	key2 := "key2"
	val1 := []byte("val1")
	val2 := []byte("val2")

	m := map[string][]byte{
		key1: val1,
		key2: val2,
	}

	s := &storageStubs.StorerStub{}
	s.GetCalled = func(key []byte) (bytes []byte, e error) {
		return m[string(key)], nil
	}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(1, s)
	t1, err := b.GetAll(1, [][]byte{[]byte(key1)})
	assert.Nil(t, err)
	assert.Equal(t, map[string][]byte{key1: val1}, t1)

	t2, err := b.GetAll(1, [][]byte{[]byte(key1), []byte(key2)})
	assert.Nil(t, err)
	assert.Equal(t, map[string][]byte{key1: val1, key2: val2}, t2)
}

func TestDestroy_ErrorsWhenStorerDestroyErrors(t *testing.T) {
	t.Parallel()

	s := &storageStubs.StorerStub{}
	destroyError := errors.New("error")
	s.DestroyUnitCalled = func() error {
		return destroyError
	}
	b := dataRetriever.NewChainStorer()
	b.AddStorer(1, s)
	err := b.Destroy()
	assert.Equal(t, destroyError, err)
}

func TestDestroy_ReturnsCorrectly(t *testing.T) {
	t.Parallel()

	s := &storageStubs.StorerStub{}
	destroyCalled := false
	s.DestroyUnitCalled = func() error {
		destroyCalled = true
		return nil
	}
	b := dataRetriever.NewChainStorer()

	b.AddStorer(1, s)
	err := b.Destroy()

	assert.Nil(t, err)
	assert.True(t, destroyCalled)
}

type extendedStub struct {
	*storageStubs.StorerStub
	SetEpochForPutOperationCalled func(epoch uint32)
}

func (es *extendedStub) SetEpochForPutOperation(epoch uint32) {
	es.SetEpochForPutOperationCalled(epoch)
}

func TestBlockChain_SetEpochForPutOperation(t *testing.T) {
	t.Parallel()

	expectedEpoch := uint32(37)
	setEpochWasCalled := false
	headerUnit := &extendedStub{}
	headerUnit.SetEpochForPutOperationCalled = func(epoch uint32) {
		assert.Equal(t, expectedEpoch, epoch)
		setEpochWasCalled = true
	}
	txUnit := &storageStubs.StorerStub{}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(0, headerUnit)
	b.AddStorer(1, txUnit)

	b.SetEpochForPutOperation(expectedEpoch)
	assert.True(t, setEpochWasCalled)
}

func TestBlockChain_GetStorer(t *testing.T) {
	t.Parallel()

	txUnit := &storageStubs.StorerStub{}
	txBlockUnit := &storageStubs.StorerStub{}
	stateBlockUnit := &storageStubs.StorerStub{}
	peerBlockUnit := &storageStubs.StorerStub{}
	headerUnit := &storageStubs.StorerStub{}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(0, txUnit)
	b.AddStorer(1, txBlockUnit)
	b.AddStorer(2, stateBlockUnit)
	b.AddStorer(3, peerBlockUnit)
	b.AddStorer(4, headerUnit)

	storer, _ := b.GetStorer(0)
	assert.True(t, txUnit == storer)
	storer, _ = b.GetStorer(1)
	assert.True(t, txBlockUnit == storer)
	storer, _ = b.GetStorer(2)
	assert.True(t, stateBlockUnit == storer)
	storer, _ = b.GetStorer(3)
	assert.True(t, peerBlockUnit == storer)
	storer, _ = b.GetStorer(4)
	assert.True(t, headerUnit == storer)
	storer, err := b.GetStorer(5)
	assert.True(t, errors.Is(err, dataRetriever.ErrStorerNotFound))
	assert.Nil(t, storer)
}

func TestBlockChain_GetAllStorers(t *testing.T) {
	t.Parallel()

	txUnit := &storageStubs.StorerStub{}
	txBlockUnit := &storageStubs.StorerStub{}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(0, txUnit)
	b.AddStorer(1, txBlockUnit)

	allStorers := b.GetAllStorers()
	assert.Equal(t, txUnit, allStorers[0])
	assert.Equal(t, txBlockUnit, allStorers[1])
	assert.Len(t, allStorers, 2)
}

func TestCloseAll_Error(t *testing.T) {
	t.Parallel()

	closeErr := errors.New("error")
	s := &storageStubs.StorerStub{
		CloseCalled: func() error {
			return closeErr
		},
	}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(1, s)

	err := b.CloseAll()
	require.Equal(t, storage.ErrClosingPersisters, err)
}

func TestCloseAll_Ok(t *testing.T) {
	t.Parallel()

	s := &storageStubs.StorerStub{}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(1, s)

	err := b.CloseAll()
	require.Nil(t, err)
}

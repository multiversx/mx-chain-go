package dataRetriever_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/stretchr/testify/assert"
)

func TestHas_ErrorWhenStorerIsMissing(t *testing.T) {
	s := &mock.StorerStub{}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(1, s)
	err := b.Has(2, []byte("whatever"))
	assert.Equal(t, dataRetriever.ErrNoSuchStorageUnit, err)
}

func TestHas_ReturnsCorrectly(t *testing.T) {
	s := &mock.StorerStub{}
	s.HasCalled = func(key []byte) error {
		return nil
	}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(1, s)
	err := b.Has(1, []byte("whatever"))
	assert.Nil(t, err)
}

func TestGet_ErrorWhenStorerIsMissing(t *testing.T) {
	s := &mock.StorerStub{}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(1, s)
	obj, err := b.Get(2, []byte("whatever"))
	assert.Nil(t, obj)
	assert.Equal(t, dataRetriever.ErrNoSuchStorageUnit, err)
}

func TestGet_ReturnsCorrectly(t *testing.T) {
	s := &mock.StorerStub{}
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
	s := &mock.StorerStub{}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(1, s)
	err := b.Put(2, []byte("whatever"), []byte("whatever value"))

	assert.Equal(t, dataRetriever.ErrNoSuchStorageUnit, err)
}

func TestPut_ReturnsCorrectly(t *testing.T) {
	s := &mock.StorerStub{}
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
	s := &mock.StorerStub{}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(1, s)
	ret, err := b.GetAll(2, [][]byte{[]byte("whatever"), []byte("whatever 2")})

	assert.Nil(t, ret)
	assert.Equal(t, dataRetriever.ErrNoSuchStorageUnit, err)
}

func TestGetAll_ErrorWhenStorersGetErrors(t *testing.T) {
	s := &mock.StorerStub{}
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
	key1 := "key1"
	key2 := "key2"
	val1 := []byte("val1")
	val2 := []byte("val2")

	m := map[string][]byte{
		key1: val1,
		key2: val2,
	}

	s := &mock.StorerStub{}
	s.GetCalled = func(key []byte) (bytes []byte, e error) {
		return m[string(key)], nil
	}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(1, s)
	t1, err := b.GetAll(1, [][]byte{[]byte(key1)})
	assert.Nil(t, err)
	assert.Equal(t, map[string][]byte{key1: val1}, t1)

	t2, err := b.GetAll(1, [][]byte{[]byte(key1), []byte(key2)})
	assert.Equal(t, map[string][]byte{key1: val1, key2: val2}, t2)
}

func TestDestroy_ErrrorsWhenStorerDestroyErrors(t *testing.T) {
	s := &mock.StorerStub{}
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
	s := &mock.StorerStub{}
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

func TestBlockChain_GetStorer(t *testing.T) {
	t.Parallel()

	txUnit := &mock.StorerStub{}
	txBlockUnit := &mock.StorerStub{}
	stateBlockUnit := &mock.StorerStub{}
	peerBlockUnit := &mock.StorerStub{}
	headerUnit := &mock.StorerStub{}

	b := dataRetriever.NewChainStorer()
	b.AddStorer(0, txUnit)
	b.AddStorer(1, txBlockUnit)
	b.AddStorer(2, stateBlockUnit)
	b.AddStorer(3, peerBlockUnit)
	b.AddStorer(4, headerUnit)

	assert.True(t, txUnit == b.GetStorer(0))
	assert.True(t, txBlockUnit == b.GetStorer(1))
	assert.True(t, stateBlockUnit == b.GetStorer(2))
	assert.True(t, peerBlockUnit == b.GetStorer(3))
	assert.True(t, headerUnit == b.GetStorer(4))
}

package containers_test

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

func TestNewResolversContainer_ShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	assert.False(t, check.IfNil(c))
}

//------- Add

func TestResolversContainer_AddAlreadyExistingShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	_ = c.Add("key", &mock.ResolverStub{})
	err := c.Add("key", &mock.ResolverStub{})

	assert.Equal(t, process.ErrContainerKeyAlreadyExists, err)
}

func TestResolversContainer_AddNilShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	err := c.Add("key", nil)

	assert.Equal(t, process.ErrNilContainerElement, err)
}

func TestResolversContainer_AddEmptyKeyShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	err := c.Add("", &mock.ResolverStub{})

	assert.Nil(t, err)
}

func TestResolversContainer_AddShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	err := c.Add("key", &mock.ResolverStub{})

	assert.Nil(t, err)
}

//------- AddMultiple

func TestResolversContainer_AddMultipleAlreadyExistingShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	keys := []string{"key", "key"}
	resolvers := []dataRetriever.Resolver{&mock.ResolverStub{}, &mock.ResolverStub{}}

	err := c.AddMultiple(keys, resolvers)

	assert.Equal(t, process.ErrContainerKeyAlreadyExists, err)
}

func TestResolversContainer_AddMultipleLenMismatchShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	keys := []string{"key"}
	resolvers := []dataRetriever.Resolver{&mock.ResolverStub{}, &mock.ResolverStub{}}

	err := c.AddMultiple(keys, resolvers)

	assert.Equal(t, process.ErrLenMismatch, err)
}

func TestResolversContainer_AddMultipleShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	keys := []string{"key1", "key2"}
	resolvers := []dataRetriever.Resolver{&mock.ResolverStub{}, &mock.ResolverStub{}}

	err := c.AddMultiple(keys, resolvers)

	assert.Nil(t, err)
	assert.Equal(t, 2, c.Len())
}

//------- Get

func TestResolversContainer_GetNotFoundShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	key := "key"
	keyNotFound := "key not found"
	val := &mock.ResolverStub{}

	_ = c.Add(key, val)
	valRecovered, err := c.Get(keyNotFound)

	assert.Nil(t, valRecovered)
	assert.True(t, errors.Is(err, dataRetriever.ErrInvalidContainerKey))
}

func TestResolversContainer_GetWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	key := "key"

	_ = c.Insert(key, "string value")
	valRecovered, err := c.Get(key)

	assert.Nil(t, valRecovered)
	assert.Equal(t, process.ErrWrongTypeInContainer, err)
}

func TestResolversContainer_GetShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	key := "key"
	val := &mock.ResolverStub{}

	_ = c.Add(key, val)
	valRecovered, err := c.Get(key)

	assert.True(t, val == valRecovered)
	assert.Nil(t, err)
}

//------- Replace

func TestResolversContainer_ReplaceNilValueShouldErrAndNotModify(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	key := "key"
	val := &mock.ResolverStub{}

	_ = c.Add(key, val)
	err := c.Replace(key, nil)

	valRecovered, _ := c.Get(key)

	assert.Equal(t, process.ErrNilContainerElement, err)
	assert.Equal(t, val, valRecovered)
}

func TestResolversContainer_ReplaceShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	key := "key"
	val := &mock.ResolverStub{}
	val2 := &mock.ResolverStub{}

	_ = c.Add(key, val)
	err := c.Replace(key, val2)

	valRecovered, _ := c.Get(key)

	assert.True(t, val2 == valRecovered)
	assert.Nil(t, err)
}

//------- Remove

func TestResolversContainer_RemoveShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	key := "key"
	val := &mock.ResolverStub{}

	_ = c.Add(key, val)
	c.Remove(key)

	valRecovered, err := c.Get(key)

	assert.Nil(t, valRecovered)
	assert.True(t, errors.Is(err, dataRetriever.ErrInvalidContainerKey))
}

//------- Len

func TestResolversContainer_LenShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	_ = c.Add("key1", &mock.ResolverStub{})
	assert.Equal(t, 1, c.Len())

	_ = c.Add("key2", &mock.ResolverStub{})
	assert.Equal(t, 2, c.Len())

	c.Remove("key1")
	assert.Equal(t, 1, c.Len())
}

//------- ResolverKeys

func TestResolversContainer_ResolverKeys(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	_ = c.Add("key1", &mock.ResolverStub{})
	_ = c.Add("key0", &mock.ResolverStub{})
	_ = c.Add("key2", &mock.ResolverStub{})
	_ = c.Add("a", &mock.ResolverStub{})

	expectedString := "a, key0, key1, key2"

	assert.Equal(t, expectedString, c.ResolverKeys())
}

//-------- Iterate

func TestResolversContainer_IterateNilHandlerShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have paniced")
		}
	}()

	c := containers.NewResolversContainer()

	_ = c.Add("key1", &mock.ResolverStub{})
	_ = c.Add("key2", &mock.ResolverStub{})

	c.Iterate(nil)
}

func TestResolversContainer_IterateNotAValidKeyShouldWorkAndNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have paniced")
		}
	}()

	c := containers.NewResolversContainer()

	_ = c.Add("key1", &mock.ResolverStub{})

	runs := uint32(0)
	c.Iterate(func(key string, resolver dataRetriever.Resolver) bool {
		atomic.AddUint32(&runs, 1)
		return true
	})

	assert.Equal(t, uint32(1), atomic.LoadUint32(&runs))
}

func TestResolversContainer_IterateNotAValidValueShouldWorkAndNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have paniced")
		}
	}()

	c := containers.NewResolversContainer()

	_ = c.Add("key1", &mock.ResolverStub{})
	c.Objects().Set("key 2", struct{}{})

	runs := uint32(0)
	c.Iterate(func(key string, resolver dataRetriever.Resolver) bool {
		atomic.AddUint32(&runs, 1)
		return true
	})

	assert.Equal(t, uint32(1), atomic.LoadUint32(&runs))
}

func TestResolversContainer_Iterate(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	_ = c.Add("key1", &mock.ResolverStub{})
	_ = c.Add("key2", &mock.ResolverStub{})

	runs := uint32(0)
	c.Iterate(func(key string, resolver dataRetriever.Resolver) bool {
		atomic.AddUint32(&runs, 1)
		return true
	})

	assert.Equal(t, uint32(2), atomic.LoadUint32(&runs))
}

func TestResolversContainer_IterateEarlyExitShouldWork(t *testing.T) {
	t.Parallel()

	c := containers.NewResolversContainer()

	_ = c.Add("key1", &mock.ResolverStub{})
	_ = c.Add("key2", &mock.ResolverStub{})

	runs := uint32(0)
	c.Iterate(func(key string, resolver dataRetriever.Resolver) bool {
		atomic.AddUint32(&runs, 1)
		return false
	})

	assert.Equal(t, uint32(1), atomic.LoadUint32(&runs))
}

func TestResolversContainer_Close(t *testing.T) {
	t.Parallel()

	closeCalled := uint32(0)
	expectedErr := errors.New("expected error")
	res1 := &mock.ResolverStub{
		CloseCalled: func() error {
			atomic.AddUint32(&closeCalled, 1)
			return expectedErr
		},
	}
	res2 := &mock.ResolverStub{
		CloseCalled: func() error {
			atomic.AddUint32(&closeCalled, 1)
			return nil
		},
	}

	c := containers.NewResolversContainer()
	_ = c.Add("key1", res1)
	_ = c.Add("key2", res2)

	assert.Equal(t, expectedErr, c.Close())
	assert.Equal(t, uint32(2), atomic.LoadUint32(&closeCalled))
}

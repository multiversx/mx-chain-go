package containers_test

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/containers"
	"github.com/multiversx/mx-chain-go/process"
	dataRetrieverTests "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/assert"
)

func TestNewRequestersContainer_ShouldWork(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should have not panicked")
		}
	}()

	c := containers.NewRequestersContainer()

	assert.False(t, check.IfNil(c))
	assert.Nil(t, c.Close())
}

func TestRequestersContainer_Add(t *testing.T) {
	t.Parallel()

	t.Run("already existing should error", func(t *testing.T) {
		t.Parallel()

		c := containers.NewRequestersContainer()

		_ = c.Add("key", &dataRetrieverTests.RequesterStub{})
		err := c.Add("key", &dataRetrieverTests.RequesterStub{})

		assert.Equal(t, process.ErrContainerKeyAlreadyExists, err)
	})
	t.Run("nil requester should error", func(t *testing.T) {
		t.Parallel()

		c := containers.NewRequestersContainer()

		err := c.Add("key", nil)

		assert.Equal(t, process.ErrNilContainerElement, err)
	})
	t.Run("empty key should work", func(t *testing.T) {
		t.Parallel()

		c := containers.NewRequestersContainer()

		err := c.Add("", &dataRetrieverTests.RequesterStub{})

		assert.Nil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		c := containers.NewRequestersContainer()

		err := c.Add("key", &dataRetrieverTests.RequesterStub{})

		assert.Nil(t, err)
	})
}

func TestRequestersContainer_AddMultiple(t *testing.T) {
	t.Parallel()

	t.Run("already existing should error", func(t *testing.T) {
		t.Parallel()

		c := containers.NewRequestersContainer()

		keys := []string{"key", "key"}
		requesters := []dataRetriever.Requester{&dataRetrieverTests.RequesterStub{}, &dataRetrieverTests.RequesterStub{}}

		err := c.AddMultiple(keys, requesters)

		assert.Equal(t, process.ErrContainerKeyAlreadyExists, err)
	})
	t.Run("len mismatch should error", func(t *testing.T) {
		t.Parallel()

		c := containers.NewRequestersContainer()

		keys := []string{"key"}
		requesters := []dataRetriever.Requester{&dataRetrieverTests.RequesterStub{}, &dataRetrieverTests.RequesterStub{}}

		err := c.AddMultiple(keys, requesters)

		assert.Equal(t, process.ErrLenMismatch, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		c := containers.NewRequestersContainer()

		keys := []string{"key1", "key2"}
		requesters := []dataRetriever.Requester{&dataRetrieverTests.RequesterStub{}, &dataRetrieverTests.RequesterStub{}}

		err := c.AddMultiple(keys, requesters)

		assert.Nil(t, err)
		assert.Equal(t, 2, c.Len())
	})
}

func TestRequestersContainer_Get(t *testing.T) {
	t.Parallel()

	t.Run("key not found should error", func(t *testing.T) {
		t.Parallel()

		c := containers.NewRequestersContainer()

		key := "key"
		keyNotFound := "key not found"
		val := &dataRetrieverTests.RequesterStub{}

		_ = c.Add(key, val)
		valRecovered, err := c.Get(keyNotFound)

		assert.Nil(t, valRecovered)
		assert.True(t, errors.Is(err, dataRetriever.ErrInvalidContainerKey))
	})
	t.Run("wrong type should error", func(t *testing.T) {
		t.Parallel()

		c := containers.NewRequestersContainer()

		key := "key"

		_ = c.Insert(key, "string value")
		valRecovered, err := c.Get(key)

		assert.Nil(t, valRecovered)
		assert.Equal(t, process.ErrWrongTypeInContainer, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		c := containers.NewRequestersContainer()

		key := "key"
		val := &dataRetrieverTests.RequesterStub{}

		_ = c.Add(key, val)
		valRecovered, err := c.Get(key)

		assert.True(t, val == valRecovered)
		assert.Nil(t, err)
	})
}

func TestRequestersContainer_Replace(t *testing.T) {
	t.Parallel()

	t.Run("nil value should error", func(t *testing.T) {
		t.Parallel()

		c := containers.NewRequestersContainer()

		key := "key"
		val := &dataRetrieverTests.RequesterStub{}

		_ = c.Add(key, val)
		err := c.Replace(key, nil)

		valRecovered, _ := c.Get(key)

		assert.Equal(t, process.ErrNilContainerElement, err)
		assert.Equal(t, val, valRecovered)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		c := containers.NewRequestersContainer()

		key := "key"
		val := &dataRetrieverTests.RequesterStub{}
		val2 := &dataRetrieverTests.RequesterStub{}

		_ = c.Add(key, val)
		err := c.Replace(key, val2)

		valRecovered, _ := c.Get(key)

		assert.True(t, val2 == valRecovered)
		assert.Nil(t, err)
	})
}

func TestRequestersContainer_Remove(t *testing.T) {
	t.Parallel()

	c := containers.NewRequestersContainer()

	key := "key"
	val := &dataRetrieverTests.RequesterStub{}

	_ = c.Add(key, val)
	c.Remove(key)

	valRecovered, err := c.Get(key)

	assert.Nil(t, valRecovered)
	assert.True(t, errors.Is(err, dataRetriever.ErrInvalidContainerKey))
}

func TestRequestersContainer_Len(t *testing.T) {
	t.Parallel()

	c := containers.NewRequestersContainer()

	_ = c.Add("key1", &dataRetrieverTests.RequesterStub{})
	assert.Equal(t, 1, c.Len())

	_ = c.Add("key2", &dataRetrieverTests.RequesterStub{})
	assert.Equal(t, 2, c.Len())

	c.Remove("key1")
	assert.Equal(t, 1, c.Len())
}

func TestRequestersContainer_RequesterKeys(t *testing.T) {
	t.Parallel()

	c := containers.NewRequestersContainer()

	_ = c.Add("key1", &dataRetrieverTests.RequesterStub{})
	_ = c.Add("key0", &dataRetrieverTests.RequesterStub{})
	_ = c.Add("key2", &dataRetrieverTests.RequesterStub{})
	_ = c.Add("a", &dataRetrieverTests.RequesterStub{})
	c.Objects().Insert(1234, &dataRetrieverTests.RequesterStub{}) // key not string should be skipped

	expectedString := "a, key0, key1, key2"

	assert.Equal(t, expectedString, c.RequesterKeys())
}

func TestRequestersContainer_Iterate(t *testing.T) {
	t.Parallel()

	t.Run("nil handler should not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not have panicked")
			}
		}()

		c := containers.NewRequestersContainer()

		_ = c.Add("key1", &dataRetrieverTests.RequesterStub{})
		_ = c.Add("key2", &dataRetrieverTests.RequesterStub{})

		c.Iterate(nil)
	})
	t.Run("early exist should work", func(t *testing.T) {
		t.Parallel()

		c := containers.NewRequestersContainer()

		_ = c.Add("key1", &dataRetrieverTests.RequesterStub{})
		_ = c.Add("key2", &dataRetrieverTests.RequesterStub{})

		runs := uint32(0)
		c.Iterate(func(key string, requester dataRetriever.Requester) bool {
			atomic.AddUint32(&runs, 1)
			return false
		})

		assert.Equal(t, uint32(1), atomic.LoadUint32(&runs))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not have panicked")
			}
		}()

		c := containers.NewRequestersContainer()

		_ = c.Add("key1", &dataRetrieverTests.RequesterStub{})
		c.Objects().Insert(1234, &dataRetrieverTests.RequesterStub{}) // key not string should be skipped
		c.Objects().Set("key 2", struct{}{})                          // value not requester should be skipped

		runs := uint32(0)
		c.Iterate(func(key string, requester dataRetriever.Requester) bool {
			atomic.AddUint32(&runs, 1)
			return true
		})

		assert.Equal(t, uint32(1), atomic.LoadUint32(&runs))
	})
}

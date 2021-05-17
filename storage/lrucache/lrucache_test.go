package lrucache_test

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/stretchr/testify/assert"
)

var timeoutWaitForWaitGroups = time.Second * 2

//------- NewCache

func TestNewCache_BadSizeShouldErr(t *testing.T) {
	t.Parallel()

	c, err := lrucache.NewCache(0)

	assert.True(t, check.IfNil(c))
	assert.NotNil(t, err)
}

func TestNewCache_ShouldWork(t *testing.T) {
	t.Parallel()

	c, err := lrucache.NewCache(1)

	assert.False(t, check.IfNil(c))
	assert.Nil(t, err)
}

//------- NewCacheWithSizeInBytes

func TestNewCacheWithSizeInBytes_BadSizeShouldErr(t *testing.T) {
	t.Parallel()

	c, err := lrucache.NewCacheWithSizeInBytes(0, 100000)

	assert.True(t, check.IfNil(c))
	assert.Equal(t, storage.ErrCacheSizeInvalid, err)
}

func TestNewCacheWithSizeInBytes_BadSizeInBytesShouldErr(t *testing.T) {
	t.Parallel()

	c, err := lrucache.NewCacheWithSizeInBytes(1, 0)

	assert.True(t, check.IfNil(c))
	assert.Equal(t, storage.ErrCacheCapacityInvalid, err)
}

func TestNewCacheWithSizeInBytes_ShouldWork(t *testing.T) {
	t.Parallel()

	c, err := lrucache.NewCacheWithSizeInBytes(1, 100000)

	assert.False(t, check.IfNil(c))
	assert.Nil(t, err)
}

func TestLRUCache_PutNotPresent(t *testing.T) {
	t.Parallel()

	key, val := []byte("key"), []byte("value")
	c, _ := lrucache.NewCache(10)

	l := c.Len()

	assert.Zero(t, l, "cache expected to be empty")

	c.Put(key, val, 0)
	l = c.Len()

	assert.Equal(t, l, 1, "cache size expected 1 but found %d", l)
}

func TestLRUCache_PutPresent(t *testing.T) {
	t.Parallel()

	key, val := []byte("key"), []byte("value")
	c, _ := lrucache.NewCache(10)

	c.Put(key, val, 0)
	c.Put(key, val, 0)

	l := c.Len()
	assert.Equal(t, l, 1, "cache size expected 1 but found %d", l)
}

func TestLRUCache_PutPresentRewrite(t *testing.T) {
	t.Parallel()

	key := []byte("key")
	val1 := []byte("value1")
	val2 := []byte("value2")
	c, _ := lrucache.NewCache(10)

	c.Put(key, val1, 0)
	c.Put(key, val2, 0)

	l := c.Len()
	assert.Equal(t, l, 1, "cache size expected 1 but found %d", l)
	recoveredVal, has := c.Get(key)
	assert.True(t, has)
	assert.Equal(t, val2, recoveredVal)
}

func TestLRUCache_GetNotPresent(t *testing.T) {
	t.Parallel()

	key := []byte("key1")
	c, _ := lrucache.NewCache(10)

	v, ok := c.Get(key)

	assert.False(t, ok, "value %s not expected to be found", v)
}

func TestLRUCache_GetPresent(t *testing.T) {
	t.Parallel()

	key, val := []byte("key2"), []byte("value2")
	c, _ := lrucache.NewCache(10)

	c.Put(key, val, 0)

	v, ok := c.Get(key)

	assert.True(t, ok, "value expected but not found")
	assert.Equal(t, val, v)
}

func TestLRUCache_HasNotPresent(t *testing.T) {
	t.Parallel()

	key := []byte("key3")
	c, _ := lrucache.NewCache(10)

	found := c.Has(key)

	assert.False(t, found, "key %s not expected to be found", key)
}

func TestLRUCache_HasPresent(t *testing.T) {
	t.Parallel()

	key, val := []byte("key4"), []byte("value4")
	c, _ := lrucache.NewCache(10)

	c.Put(key, val, 0)

	found := c.Has(key)

	assert.True(t, found, "value expected but not found")
}

func TestLRUCache_PeekNotPresent(t *testing.T) {
	t.Parallel()

	key := []byte("key5")
	c, _ := lrucache.NewCache(10)

	_, ok := c.Peek(key)

	assert.False(t, ok, "not expected to find key %s", key)
}

func TestLRUCache_PeekPresent(t *testing.T) {
	t.Parallel()

	key, val := []byte("key6"), []byte("value6")
	c, _ := lrucache.NewCache(10)

	c.Put(key, val, 0)
	v, ok := c.Peek(key)

	assert.True(t, ok, "value expected but not found")
	assert.Equal(t, val, v, "expected to find %s but found %s", val, v)
}

func TestLRUCache_HasOrAddNotPresent(t *testing.T) {
	t.Parallel()

	key, val := []byte("key7"), []byte("value7")
	c, _ := lrucache.NewCache(10)

	_, ok := c.Peek(key)
	assert.False(t, ok, "not expected to find key %s", key)

	c.HasOrAdd(key, val, 0)
	v, ok := c.Peek(key)
	assert.True(t, ok, "value expected but not found")
	assert.Equal(t, val, v, "expected to find %s but found %s", val, v)
}

func TestLRUCache_HasOrAddPresent(t *testing.T) {
	t.Parallel()

	key, val := []byte("key8"), []byte("value8")
	c, _ := lrucache.NewCache(10)

	_, ok := c.Peek(key)

	assert.False(t, ok, "not expected to find key %s", key)

	c.HasOrAdd(key, val, 0)
	v, ok := c.Peek(key)

	assert.True(t, ok, "value expected but not found")
	assert.Equal(t, val, v, "expected to find %s but found %s", val, v)
}

func TestLRUCache_RemoveNotPresent(t *testing.T) {
	t.Parallel()

	key := []byte("key9")
	c, _ := lrucache.NewCache(10)

	found := c.Has(key)

	assert.False(t, found, "not expected to find key %s", key)

	c.Remove(key)
	found = c.Has(key)

	assert.False(t, found, "not expected to find key %s", key)
}

func TestLRUCache_RemovePresent(t *testing.T) {
	t.Parallel()

	key, val := []byte("key10"), []byte("value10")
	c, _ := lrucache.NewCache(10)

	c.Put(key, val, 0)
	found := c.Has(key)

	assert.True(t, found, "expected to find key %s", key)

	c.Remove(key)
	found = c.Has(key)

	assert.False(t, found, "not expected to find key %s", key)
}

func TestLRUCache_Keys(t *testing.T) {
	t.Parallel()

	c, _ := lrucache.NewCache(10)

	for i := 0; i < 20; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Put(key, val, 0)
	}

	keys := c.Keys()

	// check also that cache size does not grow over the capacity
	assert.Equal(t, 10, len(keys), "expected cache size 10 but current size %d", len(keys))
}

func TestLRUCache_Len(t *testing.T) {
	t.Parallel()

	c, _ := lrucache.NewCache(10)

	for i := 0; i < 20; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Put(key, val, 0)
	}

	l := c.Len()

	assert.Equal(t, 10, l, "expected cache size 10 but current size %d", l)
}

func TestLRUCache_Clear(t *testing.T) {
	t.Parallel()

	c, _ := lrucache.NewCache(10)

	for i := 0; i < 5; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Put(key, val, 0)
	}

	l := c.Len()

	assert.Equal(t, 5, l, "expected size 5, got %d", l)

	c.Clear()
	l = c.Len()

	assert.Zero(t, l, "expected size 0, got %d", l)
}

func TestLRUCache_CacherRegisterAddedDataHandlerNilHandlerShouldIgnore(t *testing.T) {
	t.Parallel()

	c, _ := lrucache.NewCache(100)
	c.RegisterHandler(nil, "")

	assert.Equal(t, 0, len(c.AddedDataHandlers()))
}

func TestLRUCache_CacherRegisterPutAddedDataHandlerShouldWork(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	chDone := make(chan bool)

	f := func(key []byte, value interface{}) {
		if !bytes.Equal([]byte("aaaa"), key) {
			return
		}

		wg.Done()
	}

	go func() {
		wg.Wait()
		chDone <- true
	}()

	c, _ := lrucache.NewCache(100)
	c.RegisterHandler(f, "")
	c.Put([]byte("aaaa"), "bbbb", 0)

	select {
	case <-chDone:
	case <-time.After(timeoutWaitForWaitGroups):
		assert.Fail(t, "should have been called")
		return
	}

	assert.Equal(t, 1, len(c.AddedDataHandlers()))
}

func TestLRUCache_CacherRegisterHasOrAddAddedDataHandlerShouldWork(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	chDone := make(chan bool)

	f := func(key []byte, value interface{}) {
		if !bytes.Equal([]byte("aaaa"), key) {
			return
		}

		wg.Done()
	}

	go func() {
		wg.Wait()
		chDone <- true
	}()

	c, _ := lrucache.NewCache(100)
	c.RegisterHandler(f, "")
	c.HasOrAdd([]byte("aaaa"), "bbbb", 0)

	select {
	case <-chDone:
	case <-time.After(timeoutWaitForWaitGroups):
		assert.Fail(t, "should have been called")
		return
	}

	assert.Equal(t, 1, len(c.AddedDataHandlers()))
}

func TestLRUCache_CacherRegisterHasOrAddAddedDataHandlerNotAddedShouldNotCall(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	chDone := make(chan bool)

	f := func(key []byte, value interface{}) {
		wg.Done()
	}

	go func() {
		wg.Wait()
		chDone <- true
	}()

	c, _ := lrucache.NewCache(100)
	//first add, no call
	c.HasOrAdd([]byte("aaaa"), "bbbb", 0)
	c.RegisterHandler(f, "")
	//second add, should not call as the data was found
	c.HasOrAdd([]byte("aaaa"), "bbbb", 0)

	select {
	case <-chDone:
		assert.Fail(t, "should have not been called")
		return
	case <-time.After(timeoutWaitForWaitGroups):
	}

	assert.Equal(t, 1, len(c.AddedDataHandlers()))
}

func TestLRUCache_CloseShouldNotErr(t *testing.T) {
	t.Parallel()

	c, _ := lrucache.NewCache(1)

	err := c.Close()
	assert.Nil(t, err)
}

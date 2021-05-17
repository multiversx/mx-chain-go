package fifocache_test

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/storage/fifocache"
	"github.com/stretchr/testify/assert"
)

var timeoutWaitForWaitGroups = time.Second * 2

func TestFIFOShardedCache_PutNotPresent(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	c, err := fifocache.NewShardedCache(10, 2)

	assert.Nil(t, err, "no error expected but got %s", err)

	l := c.Len()

	assert.Zero(t, l, "cache expected to be empty")

	c.Put(key, val, 0)
	l = c.Len()

	assert.Equal(t, l, 1, "cache size expected 1 but found %d", l)
}

func TestFIFOShardedCache_PutPresent(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	c, err := fifocache.NewShardedCache(10, 2)

	assert.Nil(t, err, "no error expected but got %s", err)

	c.Put(key, val, 0)
	c.Put(key, val, 0)

	l := c.Len()
	assert.Equal(t, l, 1, "cache size expected 1 but found %d", l)
}

func TestFIFOShardedCache_PutPresentRewrite(t *testing.T) {
	key := []byte("key")
	val1 := []byte("value1")
	val2 := []byte("value2")
	c, err := fifocache.NewShardedCache(10, 2)

	assert.Nil(t, err, "no error expected but got %s", err)

	c.Put(key, val1, 0)
	c.Put(key, val2, 0)

	l := c.Len()
	assert.Equal(t, l, 1, "cache size expected 1 but found %d", l)
	recoveredVal, has := c.Get(key)
	assert.True(t, has)
	assert.Equal(t, val2, recoveredVal)
}

func TestFIFOShardedCache_GetNotPresent(t *testing.T) {
	key := []byte("key1")
	c, err := fifocache.NewShardedCache(10, 2)

	assert.Nil(t, err, "no error expected but got %s", err)

	v, ok := c.Get(key)

	assert.False(t, ok, "value %s not expected to be found", v)
}

func TestFIFOShardedCache_GetPresent(t *testing.T) {
	key, val := []byte("key2"), []byte("value2")
	c, err := fifocache.NewShardedCache(10, 2)

	assert.Nil(t, err, "no error expected but got %s", err)

	c.Put(key, val, 0)

	v, ok := c.Get(key)

	assert.True(t, ok, "value expected but not found")
	assert.Equal(t, val, v)
}

func TestFIFOShardedCache_HasNotPresent(t *testing.T) {
	key := []byte("key3")
	c, err := fifocache.NewShardedCache(10, 2)

	assert.Nil(t, err, "no error expected but got %s", err)

	found := c.Has(key)

	assert.False(t, found, "key %s not expected to be found", key)
}

func TestFIFOShardedCache_HasPresent(t *testing.T) {
	key, val := []byte("key4"), []byte("value4")
	c, err := fifocache.NewShardedCache(10, 2)

	assert.Nil(t, err, "no error expected but got %s", err)

	c.Put(key, val, 0)

	found := c.Has(key)

	assert.True(t, found, "value expected but not found")
}

func TestFIFOShardedCache_PeekNotPresent(t *testing.T) {
	key := []byte("key5")
	c, err := fifocache.NewShardedCache(10, 2)

	assert.Nil(t, err, "no error expected but got %s", err)

	_, ok := c.Peek(key)

	assert.False(t, ok, "not expected to find key %s", key)
}

func TestFIFOShardedCache_PeekPresent(t *testing.T) {
	key, val := []byte("key6"), []byte("value6")
	c, err := fifocache.NewShardedCache(10, 2)

	assert.Nil(t, err, "no error expected but got %s", err)

	c.Put(key, val, 0)
	v, ok := c.Peek(key)

	assert.True(t, ok, "value expected but not found")
	assert.Equal(t, val, v, "expected to find %s but found %s", val, v)
}

func TestFIFOShardedCache_HasOrAddNotPresent(t *testing.T) {
	key, val := []byte("key7"), []byte("value7")
	c, err := fifocache.NewShardedCache(10, 2)

	assert.Nil(t, err, "no error expected but got %s", err)

	_, ok := c.Peek(key)

	assert.False(t, ok, "not expected to find key %s", key)

	c.HasOrAdd(key, val, 0)
	v, ok := c.Peek(key)

	assert.True(t, ok, "value expected but not found")
	assert.Equal(t, val, v, "expected to find %s but found %s", val, v)
}

func TestFIFOShardedCache_HasOrAddPresent(t *testing.T) {
	key, val := []byte("key8"), []byte("value8")
	c, err := fifocache.NewShardedCache(10, 2)

	assert.Nil(t, err, "no error expected but got %s", err)

	_, ok := c.Peek(key)

	assert.False(t, ok, "not expected to find key %s", key)

	c.HasOrAdd(key, val, 0)
	v, ok := c.Peek(key)

	assert.True(t, ok, "value expected but not found")
	assert.Equal(t, val, v, "expected to find %s but found %s", val, v)
}

func TestFIFOShardedCache_RemoveNotPresent(t *testing.T) {
	key := []byte("key9")
	c, err := fifocache.NewShardedCache(10, 2)

	assert.Nil(t, err, "no error expected but got %s", err)

	found := c.Has(key)

	assert.False(t, found, "not expected to find key %s", key)

	c.Remove(key)
	found = c.Has(key)

	assert.False(t, found, "not expected to find key %s", key)
}

func TestFIFOShardedCache_RemovePresent(t *testing.T) {
	key, val := []byte("key10"), []byte("value10")
	c, err := fifocache.NewShardedCache(10, 2)

	assert.Nil(t, err, "no error expected but got %s", err)

	c.Put(key, val, 0)
	found := c.Has(key)

	assert.True(t, found, "expected to find key %s", key)

	c.Remove(key)
	found = c.Has(key)

	assert.False(t, found, "not expected to find key %s", key)
}

func TestFIFOShardedCache_Keys(t *testing.T) {
	c, err := fifocache.NewShardedCache(10, 2)

	assert.Nil(t, err, "no error expected but got %s", err)

	for i := 0; i < 20; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Put(key, val, 0)
	}

	keys := c.Keys()

	// check also that cache size does not grow over the capacity
	assert.True(t, 10 >= len(keys), "expected up to 10 stored keys but current found %d", len(keys))
}

func TestFIFOShardedCache_Len(t *testing.T) {
	c, err := fifocache.NewShardedCache(10, 2)

	assert.Nil(t, err, "no error expected but got %s", err)

	for i := 0; i < 20; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Put(key, val, 0)
	}

	l := c.Len()

	assert.True(t, 10 >= l, "expected up to 10 stored keys but current size %d", l)
}

func TestFIFOShardedCache_Clear(t *testing.T) {
	c, err := fifocache.NewShardedCache(10, 2)

	assert.Nil(t, err, "no error expected but got %s", err)

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

func TestFIFOShardedCache_CloseDoesNotErr(t *testing.T) {
	t.Parallel()

	c, _ := fifocache.NewShardedCache(10, 2)

	err := c.Close()
	assert.Nil(t, err)
}

func TestFIFOShardedCache_CacherRegisterAddedDataHandlerNilHandlerShouldIgnore(t *testing.T) {
	t.Parallel()

	c, err := fifocache.NewShardedCache(100, 2)
	assert.Nil(t, err)
	c.RegisterHandler(nil, "")

	assert.Equal(t, 0, len(c.AddedDataHandlers()))
}

func TestFIFOShardedCache_CacherRegisterPutAddedDataHandlerShouldWork(t *testing.T) {
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

	c, err := fifocache.NewShardedCache(100, 2)
	assert.Nil(t, err)
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

func TestFIFOShardedCache_CacherRegisterHasOrAddAddedDataHandlerShouldWork(t *testing.T) {
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

	c, err := fifocache.NewShardedCache(100, 2)
	assert.Nil(t, err)
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

func TestFIFOShardedCache_CacherRegisterHasOrAddAddedDataHandlerNotAddedShouldNotCall(t *testing.T) {
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

	c, err := fifocache.NewShardedCache(100, 2)
	assert.Nil(t, err)
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

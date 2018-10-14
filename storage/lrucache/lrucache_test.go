package lrucache_test

import (
	"fmt"
	"testing"

	"ElrondNetwork/elrond-go-sandbox/storage/lrucache"

	"github.com/stretchr/testify/assert"
)

func TestAddNotPresent(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	c, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	l := c.Len()

	assert.Zero(t, l, "cache expected to be empty")

	c.Put(key, val)
	l = c.Len()

	assert.Equal(t, l, 1, "cachhe size expected 1 but found %d", l)
}

func TestAddPresent(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	c, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	c.Put(key, val)
	c.Put(key, val)

	l := c.Len()
	assert.Equal(t, l, 1, "cache size expected 1 but found %d", l)
}

func TestGetNotPresent(t *testing.T) {
	key := []byte("key1")
	c, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	v, ok := c.Get(key)

	assert.False(t, ok, "value %s not expected to be found", v)
}

func TestGetPresent(t *testing.T) {
	key, val := []byte("key2"), []byte("value2")
	c, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	c.Put(key, val)

	v, ok := c.Get(key)

	assert.True(t, ok, "value expected but not found")
	assert.Equal(t, val, v)
}

func TestHasNotPresent(t *testing.T) {
	key := []byte("key3")
	c, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	found := c.Has(key)

	assert.False(t, found, "key %s not expected to be found", key)
}

func TestHasPresent(t *testing.T) {
	key, val := []byte("key4"), []byte("value4")
	c, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	c.Put(key, val)

	found := c.Has(key)

	assert.True(t, found, "value expected but not found")
}

func TestPeekNotPresent(t *testing.T) {
	key := []byte("key5")
	c, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	_, ok := c.Peek(key)

	assert.False(t, ok, "not expected to find key %s", key)
}

func TestPeekPresent(t *testing.T) {
	key, val := []byte("key6"), []byte("value6")
	c, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	c.Put(key, val)
	v, ok := c.Peek(key)

	assert.True(t, ok, "value expected but not found")
	assert.Equal(t, val, v, "expected to find %s but found %s", val, v)
}

func TestHasOrAddNotPresent(t *testing.T) {
	key, val := []byte("key7"), []byte("value7")
	c, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	v, ok := c.Peek(key)

	assert.False(t, ok, "not expected to find key %s", key)

	c.HasOrAdd(key, val)
	v, ok = c.Peek(key)

	assert.True(t, ok, "value expected but not found")
	assert.Equal(t, val, v, "expected to find %s but found %s", val, v)
}

func TestHasOrAddPresent(t *testing.T) {
	key, val := []byte("key8"), []byte("value8")
	c, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	v, ok := c.Peek(key)

	assert.False(t, ok, "not expected to find key %s", key)

	c.HasOrAdd(key, val)
	v, ok = c.Peek(key)

	assert.True(t, ok, "value expected but not found")
	assert.Equal(t, val, v, "expected to find %s but found %s", val, v)
}

func TestRemoveNotPresent(t *testing.T) {
	key := []byte("key9")
	c, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	found := c.Has(key)

	assert.False(t, found, "not expected to find key %s", key)

	c.Remove(key)
	found = c.Has(key)

	assert.False(t, found, "not expected to find key %s", key)
}

func TestRemovePresent(t *testing.T) {
	key, val := []byte("key10"), []byte("value10")
	c, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	c.Put(key, val)
	found := c.Has(key)

	assert.True(t, found, "expected to find key %s", key)

	c.Remove(key)
	found = c.Has(key)

	assert.False(t, found, "not expected to find key %s", key)
}

func TestRemoveOldestEmpty(t *testing.T) {
	c, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	l := c.Len()

	assert.Zero(t, l, "expected size 0 but got %d", l)

	c.RemoveOldest()

	l = c.Len()

	assert.Zero(t, l, "expected size 0 but got %d", l)
}

func TestRemoveOldestPresent(t *testing.T) {
	c, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	for i := 0; i < 5; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Put(key, val)
	}

	l := c.Len()

	assert.Equal(t, l, 5, "expected 5 elements in cache but found %d", l)
	c.RemoveOldest()
	found := c.Has([]byte("key0"))

	assert.False(t, found, "not expected to find key key0")
}

func TestKeys(t *testing.T) {
	c, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	for i := 0; i < 20; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Put(key, val)
	}

	keys := c.Keys()

	// check also that cache size does not grow over the capacity
	assert.Equal(t, 10, len(keys), "expected cache size 10 but current size %d", len(keys))
}

func TestLen(t *testing.T) {
	c, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	for i := 0; i < 20; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Put(key, val)
	}

	l := c.Len()

	assert.Equal(t, 10, l, "expected cache size 10 but current size %d", l)
}

func TestClear(t *testing.T) {
	c, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	for i := 0; i < 5; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Put(key, val)
	}

	l := c.Len()

	assert.Equal(t, 5, l, "expected size 5, got %d", l)

	c.Clear()
	l = c.Len()

	assert.Zero(t, l, "expected size 0, got %d", l)
}

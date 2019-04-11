package ccache_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/storage/ccache"
	"github.com/stretchr/testify/assert"
)

func TestCCache_Clear(t *testing.T) {
	c, err := ccache.NewCCache(10)

	assert.Nil(t, err, "no error expected, but got %v", err)

	for i := 0; i < 10; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Put(key, val)
	}

	assert.Equal(t, 10, c.Len(), "expected map size: 10, got %d", c.Len())

	c.Clear()

	assert.Zero(t, 0, "expected map size: 0, got %d", c.Len())
}

func TestCCache_PutNotInvoked(t *testing.T) {
	c, err := ccache.NewCCache(10)

	assert.Nil(t, err, "no error expected, but got %v", err)

	l := c.Len()

	assert.Zero(t, l, "cache expected to be empty")
}

func TestCCache_PutInvoked(t *testing.T) {
	c, err := ccache.NewCCache(10)

	assert.Nil(t, err, "no error expected, but got %v", err)

	key, val := []byte("key"), []byte("value")

	l := c.Len()
	assert.Zero(t, l, "cache expected to be empty")

	c.Put(key, val)

	l = c.Len()

	assert.Equal(t, 1, l, "cache size expected to be 1, got %d", l)
}

func TestCCache_PutRewrite(t *testing.T) {
	c, err := ccache.NewCCache(10)

	assert.Nil(t, err, "no error expected, but got %v", err)

	key := []byte("key")
	val1, val2 := []byte("value1"), []byte("value2")

	l := c.Len()
	assert.Zero(t, l, "cache expected to be empty")

	c.Put(key, val1)
	c.Put(key, val2)

	l = c.Len()

	newVal, ok := c.Get(key)
	assert.True(t, ok)

	assert.Equal(t, val2, newVal, "key new value should be %s, got %s", val2, newVal)
}

func TestCCache_GetNotPresent(t *testing.T) {
	c, err := ccache.NewCCache(10)

	assert.Nil(t, err, "no error expected, but got %v", err)

	key := []byte("key")

	val, ok := c.Get(key)

	assert.False(t, ok, "value %v was not expected to be found", val)
}

func TestCCache_GetPresent(t *testing.T) {
	c, err := ccache.NewCCache(10)

	assert.Nil(t, err, "no error expected, but got %v", err)

	key, val := []byte("key"), []byte("val")

	l := c.Len()
	assert.Zero(t, l, "cache expected to be empty")

	c.Put(key, val)

	v, ok := c.Get(key)

	assert.True(t, ok, "value %s was expected to be found", val)
	assert.Equal(t, val, v, "value %s was expected to be found, got %s", val, v)
}

func TestCCache_HasNotPresent(t *testing.T) {
	c, err := ccache.NewCCache(10)

	assert.Nil(t, err, "no error expected, but got %v", err)

	key := []byte("key")

	ok := c.Has(key)

	assert.False(t, ok, "key was not expected to be found")
}

func TestCCache_HasPresent(t *testing.T) {
	c, err := ccache.NewCCache(10)

	assert.Nil(t, err, "no error expected, but got %v", err)

	key, val := []byte("key"), []byte("val")

	l := c.Len()
	assert.Zero(t, l, "cache expected to be empty")

	c.Put(key, val)

	ok := c.Has(key)

	assert.True(t, ok, "key was expected to be found")
}

func TestCCache_PeekNotPresent(t *testing.T) {
	c, err := ccache.NewCCache(10)

	assert.Nil(t, err, "no error expected, but got %v", err)

	key := []byte("key")

	val, ok := c.Peek(key)

	assert.False(t, ok, "value %v was not expected to be found", val)
}

func TestCCache_PeekPresent(t *testing.T) {
	c, err := ccache.NewCCache(10)

	assert.Nil(t, err, "no error expected, but got %v", err)

	key, val := []byte("key"), []byte("val")

	l := c.Len()
	assert.Zero(t, l, "cache expected to be empty")

	c.Put(key, val)

	v, ok := c.Peek(key)

	assert.True(t, ok, "value %s was expected to be found", val)
	assert.Equal(t, val, v, "value %s was expected to be found, got %s", val, v)
}

func TestCCache_RemoveNotPresent(t *testing.T) {
	c, err := ccache.NewCCache(10)

	assert.Nil(t, err, "no error expected, but got %v", err)

	key := "key2"
	found := c.Has([]byte(key))

	assert.False(t, found, "not expected to find a key %s", key)

	c.Remove([]byte(key))
	found = c.Has([]byte(key))

	assert.False(t, found, "not expected to find a key %s", key)
}

func TestCCache_RemovePresent(t *testing.T) {
	c, err := ccache.NewCCache(10)

	assert.Nil(t, err, "no error expected, but got %v", err)

	for i := 0; i < 10; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Put(key, val)
	}

	key := "key2"

	c.Remove([]byte(key))

	found := c.Has([]byte(key))

	assert.False(t, found, "not expected to find a key %s", key)
}

func TestCCache_RemoveOldestPresent(t *testing.T) {
	c, err := ccache.NewCCache(10)

	assert.Nil(t, err, "no error expected, but got %v", err)

	for i := 0; i < 10; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Put(key, val)
	}

	assert.Equal(t, 10, c.Len(), "expected map size: 10, got %d", c.Len())

	oldestItem := c.FindOldest()
	c.RemoveOldest()

	found := c.Has(oldestItem)

	assert.False(t, found, "not expected to find a key %s", oldestItem)
}

func TestCCache_RemoveOldestEmpty(t *testing.T) {
	c, err := ccache.NewCCache(10)

	assert.Nil(t, err, "no error expected, but got %v", err)

	assert.Zero(t, c.Len(), "expected map size: 0, got %d", c.Len())

	c.RemoveOldest()

	assert.Zero(t, c.Len(), "expected map size: 0, got %d", c.Len())
}

func TestCCache_RemoveOldestMaxSizeExceeded(t *testing.T) {
	c, err := ccache.NewCCache(10)

	assert.Nil(t, err, "no error expected, but got %v", err)

	for i := 0; i < 20; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Put(key, val)
	}

	oldestItem := c.FindOldest()
	c.RemoveOldest()

	assert.Equal(t, 10, c.Len(), "expected map size: 10, got %d", c.Len())

	found := c.Has(oldestItem)

	assert.False(t, found, "not expected to find a key %s", oldestItem)
}

func TestCCache_Keys(t *testing.T) {
	c, err := ccache.NewCCache(10)

	assert.Nil(t, err, "no error expected but got %v", err)

	for i := 0; i < 10; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Put(key, val)
	}

	keys := c.Keys()
	assert.Equal(t, c.Len(), len(keys), "expected map size: 10, got %d", c.Len())
}

func TestCCache_Len(t *testing.T) {
	c, err := ccache.NewCCache(10)

	assert.Nil(t, err, "no error expected but got %v", err)

	for i := 0; i < 10; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Put(key, val)
	}

	assert.Equal(t, c.Len(), 10, "expected map size: 10, got %d", c.Len())
}
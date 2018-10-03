package lrucache_test

import (
	"ElrondNetwork/elrond-go-sandbox/storage"
	"ElrondNetwork/elrond-go-sandbox/storage/lrucache"
	"fmt"
	"reflect"
	"testing"
)

func TestCache(t *testing.T) {
	cache, err := lrucache.NewCache(10)

	if err != nil {
		t.Errorf("no error expected but got %s", err)
	}

	Suite(t, cache)
}

func Suite(t *testing.T, c storage.Cacher) {
	TestingAddNotPresent(t, c)
	TestingAddPresent(t, c)
	TestingGetNotPresent(t, c)
	TestingGetPresent(t, c)
	TestingContainsNotPresent(t, c)
	TestingContainsPresent(t, c)
	TestingPeekNotPresent(t, c)
	TestingPeekPresent(t, c)
	TestingContainsOrAddNotPresent(t, c)
	TestingContainsOrAddPresent(t, c)
	TestingRemoveNotPresent(t, c)
	TestingRemovePresent(t, c)
	TestingRemoveOldestEmpty(t, c)
	TestingRemoveOldestPresent(t, c)
	TestingKeys(t, c)
	TestingLen(t, c)
	TestingClear(t, c)
}

func TestingAddNotPresent(t *testing.T, c storage.Cacher) {
	key, val := []byte("key"), []byte("value")
	l := c.Len()
	if l != 0 {
		t.Errorf("cache expected to be empty")
	}
	c.Add(key, val)
	l = c.Len()
	if l != 1 {
		t.Errorf("expecteed cache size 1")
	}
}

func TestingAddPresent(t *testing.T, c storage.Cacher) {
	key, val := []byte("key"), []byte("value")

	c.Add(key, val)
	c.Add(key, val)

	l := c.Len()
	if l != 1 {
		t.Errorf("cache size expected 1 but found %d", l)
	}
}

func TestingGetNotPresent(t *testing.T, c storage.Cacher) {
	key := []byte("key1")

	v, ok := c.Get(key)

	if ok {
		t.Errorf("value %s not expected to be found", v)
	}
}

func TestingGetPresent(t *testing.T, c storage.Cacher) {
	key, val := []byte("key2"), []byte("value2")

	c.Add(key, val)

	v, ok := c.Get(key)

	if !ok {
		t.Errorf("value expected but not found")
	}

	if !reflect.DeepEqual(v, val) {
		t.Errorf("expected %s but got %s", val, v)
	}
}

func TestingContainsNotPresent(t *testing.T, c storage.Cacher) {
	key := []byte("key3")

	found := c.Contains(key)

	if found {
		t.Errorf("key %s not expected to be found", key)
	}
}

func TestingContainsPresent(t *testing.T, c storage.Cacher) {
	key, val := []byte("key4"), []byte("value4")

	c.Add(key, val)

	found := c.Contains(key)

	if !found {
		t.Errorf("value expected but not found")
	}
}

func TestingPeekNotPresent(t *testing.T, c storage.Cacher) {
	key := []byte("key5")

	_, ok := c.Peek(key)

	if ok {
		t.Errorf("not expected to find key %s", key)
	}
}

func TestingPeekPresent(t *testing.T, c storage.Cacher) {
	key, val := []byte("key6"), []byte("value6")

	c.Add(key, val)
	v, ok := c.Peek(key)

	if !ok {
		t.Errorf("value expected but not found")
	}

	if !reflect.DeepEqual(v, val) {
		t.Errorf("expected to find %s but found %s", val, v)
	}
}

func TestingContainsOrAddNotPresent(t *testing.T, c storage.Cacher) {
	key, val := []byte("key7"), []byte("value7")

	v, ok := c.Peek(key)
	if ok {
		t.Errorf("not expected to find key %s", key)
	}

	c.ContainsOrAdd(key, val)
	v, ok = c.Peek(key)

	if !ok {
		t.Errorf("value expected but not found")
	}

	if !reflect.DeepEqual(v, val) {
		t.Errorf("expected to find %s but found %s", val, v)
	}
}

func TestingContainsOrAddPresent(t *testing.T, c storage.Cacher) {
	key, val := []byte("key8"), []byte("value8")

	v, ok := c.Peek(key)
	if ok {
		t.Errorf("not expected to find key %s", key)
	}

	c.ContainsOrAdd(key, val)
	v, ok = c.Peek(key)

	if !ok {
		t.Errorf("value expected but not found")
	}

	if !reflect.DeepEqual(v, val) {
		t.Errorf("expected to find %s but found %s", val, v)
	}
}

func TestingRemoveNotPresent(t *testing.T, c storage.Cacher) {
	key := []byte("key9")
	found := c.Contains(key)

	if found {
		t.Errorf("not expected to find key %s", key)
	}

	c.Remove(key)
	found = c.Contains(key)

	if found {
		t.Errorf("not expected to find key %s", key)
	}
}

func TestingRemovePresent(t *testing.T, c storage.Cacher) {
	key, val := []byte("key10"), []byte("value10")

	c.Add(key, val)
	found := c.Contains(key)
	if !found {
		t.Errorf("expected to find key %s", key)
	}
	c.Remove(key)
	found = c.Contains(key)

	if found {
		t.Errorf("not expected to find key %s", key)
	}
}

func TestingRemoveOldestEmpty(t *testing.T, c storage.Cacher) {
	c.Clear()
	l := c.Len()
	if l != 0 {
		t.Errorf("expected size 0 but got %d", l)
	}

	c.RemoveOldest()

	l = c.Len()
	if l != 0 {
		t.Errorf("expected size 0 but got %d", l)
	}
}

func TestingRemoveOldestPresent(t *testing.T, c storage.Cacher) {
	c.Clear()
	for i := 0; i < 5; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Add(key, val)
	}
	l := c.Len()

	if l != 5 {
		t.Errorf("expected 5 elements in cache but found %d", l)
	}

	c.RemoveOldest()

	found := c.Contains([]byte("key0"))

	if found {
		t.Errorf("not expected to find key key0")
	}
}

func TestingKeys(t *testing.T, c storage.Cacher) {
	c.Clear()
	for i := 0; i < 20; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Add(key, val)
	}

	keys := c.Keys()

	// check also that cache size does not grow over the capacity
	if len(keys) != 10 {
		t.Errorf("expected cache size 10 but current size %d", len(keys))
	}
}

func TestingLen(t *testing.T, c storage.Cacher) {
	c.Clear()

	for i := 0; i < 20; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Add(key, val)
	}

}

func TestingClear(t *testing.T, c storage.Cacher) {
	c.Clear()
	for i := 0; i < 5; i++ {
		key, val := []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))
		c.Add(key, val)
	}

	l := c.Len()

	if l != 5 {
		t.Errorf("expected size 5, got %d", l)
	}

	c.Clear()

	l = c.Len()
	if l != 0 {
		t.Errorf("expected size 0, got %d", l)
	}
}

package ccache_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/fnv"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/bloom"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/ccache"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/leveldb"
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

func TestCCache_PutConcurrent(t *testing.T) {
	t.Parallel()

	const iterations = 10000
	c, err := ccache.NewCCache(iterations)

	assert.Nil(t, err, "no error expected but got %v", err)

	ch := make(chan int)
	var arr [iterations]int

	// Using go routines insert 10000 ints into our map.
	go func() {
		for i := 0; i < iterations/2; i++ {
			c.Put([]byte(strconv.Itoa(i)), i)

			val, _ := c.Get([]byte(strconv.Itoa(i)))

			ch <- val.(int)
		}
	}()

	go func() {
		for i := iterations / 2; i < iterations; i++ {
			c.Put([]byte(strconv.Itoa(i)), i)

			val, _ := c.Get([]byte(strconv.Itoa(i)))

			ch <- val.(int)
		}
	}()

	// Wait for all go routines to finish.
	idx := 0
	for elem := range ch {
		arr[idx] = elem
		idx++
		if idx == iterations {
			break
		}
	}

	// Sorts array, will make is simpler to verify all inserted values we're returned.
	sort.Ints(arr[0:iterations])

	// Make sure map contains 10000 elements.
	l := c.Len()
	assert.Equal(t, l, iterations, "expected map size: %d, got %d", iterations, l)

	// Make sure all inserted values we're fetched from map.
	for i := 0; i < iterations; i++ {
		assert.Equal(t, i, arr[i], "the value %v is missing from the map", i)
	}
}

func TestCCache_PutConcurrentWaitGroup(t *testing.T) {
	t.Parallel()

	const iterations = 10000
	c, err := ccache.NewCCache(iterations)

	assert.Nil(t, err, "no error expected but got %v", err)

	ch := make(chan int)
	done := make(chan bool)
	var arr [iterations]int

	var wg sync.WaitGroup

	wg.Add(iterations)

	// Using go routines to insert 1000 items into the map
	for i := 0; i < iterations/2; i++ {
		go func(i int) {
			defer wg.Done()
			c.Put([]byte(strconv.Itoa(i)), i)

			val, _ := c.Get([]byte(strconv.Itoa(i)))
			ch <- val.(int)
		}(i)
	}

	for i := iterations / 2; i < iterations; i++ {
		go func(i int) {
			defer wg.Done()
			c.Put([]byte(strconv.Itoa(i)), i)

			val, _ := c.Get([]byte(strconv.Itoa(i)))
			ch <- val.(int)
		}(i)
	}

	go func() {
		wg.Wait()
		close(ch)    // close the channel
		done <- true // signal completion
	}()

	idx := 0
	for elem := range ch {
		arr[idx] = elem
		idx++
	}

	// Sorts array, will make is simpler to verify all inserted values we're returned.
	sort.Ints(arr[:iterations])

	// Make sure map contains 10000 elements.
	l := c.Len()
	assert.Equal(t, l, iterations, "expected map size: %d, got %d", iterations, l)

	// Make sure all inserted values we're fetched from map.
	for i := 0; i < iterations; i++ {
		assert.Equal(t, i, arr[i], "the value %v is missing from the map", i)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		assert.Fail(t, "should have been called")
		return
	}
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

func TestCCache_RemoveConcurrent(t *testing.T) {
	t.Parallel()

	size := 1000
	c, err := ccache.NewCCache(size)

	assert.Nil(t, err, "no error expected, but got %v", err)

	var ch = make(chan int)

	// Put/Get cache value concurrently
	go func() {
		for i := 0; i < size; i++ {
			c.Put([]byte(strconv.Itoa(i)), i)

			val, _ := c.Get([]byte(strconv.Itoa(i)))

			ch <- val.(int)
		}
		close(ch) // close the channel
	}()

	wg := sync.WaitGroup{}

	// Remove cache value concurrently
	for elem := range ch {
		wg.Add(1)

		go func(elem int) {
			defer wg.Done()
			c.Remove([]byte(strconv.Itoa(elem)))
		}(elem)

		wg.Wait()
	}

	assert.Equal(t, 0, c.Len(), "expected map size: 0, got %d", c.Len())
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

func TestCCache_CacherRegisterAddedDataHandlerNilHandlerShouldIgnore(t *testing.T) {
	t.Parallel()

	c, err := ccache.NewCCache(100)
	assert.Nil(t, err)
	c.RegisterHandler(nil)

	assert.Equal(t, 0, len(c.AddedDataHandlers()))
}

func TestCCache_CacherRegisterPutAddedDataHandlerShouldWork(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	chDone := make(chan bool, 0)

	f := func(key []byte) {
		if !bytes.Equal([]byte("aaaa"), key) {
			return
		}

		wg.Done()
	}

	go func() {
		wg.Wait()
		chDone <- true
	}()

	c, err := ccache.NewCCache(100)
	assert.Nil(t, err)
	c.RegisterHandler(f)
	c.Put([]byte("aaaa"), "bbbb")

	select {
	case <-chDone:
	case <-time.After(time.Second):
		assert.Fail(t, "should have been called")
		return
	}

	assert.Equal(t, 1, len(c.AddedDataHandlers()))
}

const (
	valuesInDb = 100000
	bfSize     = 100000
)

func initSUWithNilBloomFilter(cSize int) *storage.Unit {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
	}

	ldb, err1 := leveldb.NewDB(dir + "/levelDB")
	cache, err2 := ccache.NewCCache(cSize)

	if err1 != nil {
		fmt.Println(err1)
	}
	if err2 != nil {
		fmt.Println(err2)
	}

	sUnit, err := storage.NewStorageUnit(cache, ldb)

	if err != nil {
		fmt.Println(err)
	}

	return sUnit

}

func initSUWithBloomFilter(cSize int, bfSize uint) *storage.Unit {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
	}

	ldb, err1 := leveldb.NewDB(dir + "/levelDB")
	cache, err2 := ccache.NewCCache(cSize)
	bf, err3 := bloom.NewFilter(bfSize, []hashing.Hasher{keccak.Keccak{}, blake2b.Blake2b{}, fnv.Fnv{}})

	if err1 != nil {
		fmt.Println(err1)
	}
	if err2 != nil {
		fmt.Println(err2)
	}
	if err3 != nil {
		fmt.Println(err3)
	}

	sUnit, err := storage.NewStorageUnitWithBloomFilter(cache, ldb, bf)

	if err != nil {
		fmt.Println(err)
	}

	return sUnit
}

func BenchmarkCCacheStorageUnit_PutWithNilBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithNilBloomFilter(1)
	defer func() {
		err := s.DestroyUnit()
		logError(err)
	}()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nr := rand.Intn(valuesInDb)
		b.StartTimer()

		err := s.Put([]byte(strconv.Itoa(nr)), []byte(strconv.Itoa(nr)))
		logError(err)
	}
}

func BenchmarkCCacheStorageUnit_PutWithBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithBloomFilter(100, bfSize)
	defer func() {
		err := s.DestroyUnit()
		logError(err)
	}()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nr := rand.Intn(valuesInDb)
		b.StartTimer()

		err := s.Put([]byte(strconv.Itoa(nr)), []byte(strconv.Itoa(nr)))
		logError(err)
	}
}

func BenchmarkCCacheStorageUnit_GetPresentWithNilBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithNilBloomFilter(1)
	defer func() {
		err := s.DestroyUnit()
		logError(err)
	}()
	for i := 0; i < valuesInDb; i++ {
		err := s.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
		logError(err)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nr := rand.Intn(valuesInDb)
		b.StartTimer()

		_, err := s.Get([]byte(strconv.Itoa(nr)))
		logError(err)
	}
}

func BenchmarkCCacheStorageUnit_GetPresentWithBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithBloomFilter(1, bfSize)
	defer func() {
		err := s.DestroyUnit()
		logError(err)
	}()
	for i := 0; i < valuesInDb; i++ {
		err := s.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
		logError(err)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nr := rand.Intn(valuesInDb)
		b.StartTimer()

		_, err := s.Get([]byte(strconv.Itoa(nr)))
		logError(err)
	}
}

func BenchmarkCCacheStorageUnit_GetNotPresentWithNilBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithNilBloomFilter(1)
	defer func() {
		err := s.DestroyUnit()
		logError(err)
	}()
	for i := 0; i < valuesInDb; i++ {
		err := s.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
		logError(err)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nr := rand.Intn(valuesInDb) + valuesInDb
		b.StartTimer()

		s.Get([]byte(strconv.Itoa(nr)))
	}
}

func BenchmarkCCacheStorageUnit_GetNotPresentWithBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithBloomFilter(1, bfSize)
	defer func() {
		err := s.DestroyUnit()
		logError(err)
	}()
	for i := 0; i < valuesInDb; i++ {
		err := s.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
		logError(err)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nr := rand.Intn(valuesInDb) + valuesInDb
		b.StartTimer()

		s.Get([]byte(strconv.Itoa(nr)))
	}
}

func logError(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}

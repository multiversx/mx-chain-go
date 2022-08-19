package timecache_test

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/stretchr/testify/assert"
)

func createArgTimeCacher() timecache.ArgTimeCacher {
	return timecache.ArgTimeCacher{
		DefaultSpan: time.Minute,
		CacheExpiry: time.Minute,
	}
}

func createKeysVals(numOfPairs int) ([][]byte, [][]byte) {
	keys := make([][]byte, numOfPairs)
	vals := make([][]byte, numOfPairs)
	for i := 0; i < numOfPairs; i++ {
		keys[i] = []byte("k" + string(rune(i)))
		vals[i] = []byte("v" + string(rune(i)))
	}

	return keys, vals
}

func TestNewTimeCache(t *testing.T) {
	t.Parallel()

	t.Run("invalid DefaultSpan should error", func(t *testing.T) {
		t.Parallel()

		arg := createArgTimeCacher()
		arg.DefaultSpan = time.Second - time.Nanosecond
		cacher, err := timecache.NewTimeCacher(arg)
		assert.Nil(t, cacher)
		assert.Equal(t, storage.ErrInvalidDefaultSpan, err)
	})
	t.Run("invalid CacheExpiry should error", func(t *testing.T) {
		t.Parallel()

		arg := createArgTimeCacher()
		arg.CacheExpiry = time.Second - time.Nanosecond
		cacher, err := timecache.NewTimeCacher(arg)
		assert.Nil(t, cacher)
		assert.Equal(t, storage.ErrInvalidCacheExpiry, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cacher, err := timecache.NewTimeCacher(createArgTimeCacher())
		assert.Nil(t, err)
		assert.False(t, cacher.IsInterfaceNil())
	})
}

func TestTimeCacher_Clear(t *testing.T) {
	t.Parallel()

	cacher, _ := timecache.NewTimeCacher(createArgTimeCacher())
	assert.False(t, cacher.IsInterfaceNil())

	numOfPairs := 3
	providedKeys, providedVals := createKeysVals(numOfPairs)
	for i := 0; i < numOfPairs; i++ {
		cacher.Put(providedKeys[i], providedVals[i], len(providedVals[i]))
	}
	assert.Equal(t, numOfPairs, cacher.Len())

	cacher.Clear()
	assert.Equal(t, 0, cacher.Len())
}

func TestTimeCacher_Close(t *testing.T) {
	t.Parallel()

	cacher, _ := timecache.NewTimeCacher(createArgTimeCacher())
	assert.False(t, cacher.IsInterfaceNil())

	err := cacher.Close()
	assert.Nil(t, err)
}

func TestTimeCacher_Get(t *testing.T) {
	t.Parallel()

	cacher, _ := timecache.NewTimeCacher(createArgTimeCacher())
	assert.False(t, cacher.IsInterfaceNil())

	providedKey, providedVal := []byte("key"), []byte("val")
	cacher.Put(providedKey, providedVal, len(providedVal))

	v, ok := cacher.Get(providedKey)
	assert.True(t, ok)
	assert.Equal(t, providedVal, v)

	v, ok = cacher.Get([]byte("missing key"))
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestTimeCacher_Has(t *testing.T) {
	t.Parallel()

	cacher, _ := timecache.NewTimeCacher(createArgTimeCacher())
	assert.False(t, cacher.IsInterfaceNil())

	providedKey, providedVal := []byte("key"), []byte("val")
	cacher.Put(providedKey, providedVal, len(providedVal))

	assert.True(t, cacher.Has(providedKey))
	assert.False(t, cacher.Has([]byte("missing key")))
}

func TestTimeCacher_HasOrAdd(t *testing.T) {
	t.Parallel()

	cacher, _ := timecache.NewTimeCacher(createArgTimeCacher())
	assert.False(t, cacher.IsInterfaceNil())

	providedKey, providedVal := []byte("key"), []byte("val")
	has, added := cacher.HasOrAdd(providedKey, providedVal, len(providedVal))
	assert.False(t, has)
	assert.True(t, added)

	has, added = cacher.HasOrAdd(providedKey, providedVal, len(providedVal))
	assert.True(t, has)
	assert.False(t, added)
}

func TestTimeCacher_Keys(t *testing.T) {
	t.Parallel()

	cacher, _ := timecache.NewTimeCacher(createArgTimeCacher())
	assert.False(t, cacher.IsInterfaceNil())

	numOfPairs := 10
	providedKeys, providedVals := createKeysVals(numOfPairs)
	for i := 0; i < numOfPairs; i++ {
		cacher.Put(providedKeys[i], providedVals[i], len(providedVals[i]))
	}

	receivedKeys := cacher.Keys()
	assert.Equal(t, numOfPairs, len(receivedKeys))

	sort.Slice(providedKeys, func(i, j int) bool {
		return bytes.Compare(providedKeys[i], providedKeys[j]) < 0
	})
	sort.Slice(receivedKeys, func(i, j int) bool {
		return bytes.Compare(receivedKeys[i], receivedKeys[j]) < 0
	})
	assert.Equal(t, providedKeys, receivedKeys)
}

func TestTimeCacher_Evicted(t *testing.T) {
	t.Parallel()

	arg := createArgTimeCacher()
	arg.CacheExpiry = 2 * time.Second
	arg.DefaultSpan = time.Second
	cacher, _ := timecache.NewTimeCacher(arg)
	assert.False(t, cacher.IsInterfaceNil())

	numOfPairs := 2
	providedKeys, providedVals := createKeysVals(numOfPairs)
	for i := 0; i < numOfPairs; i++ {
		cacher.Put(providedKeys[i], providedVals[i], len(providedVals[i]))
	}
	assert.Equal(t, numOfPairs, cacher.Len())

	time.Sleep(2 * arg.CacheExpiry)
	assert.Equal(t, 0, cacher.Len())
	err := cacher.Close()
	assert.Nil(t, err)
}

func TestTimeCacher_Peek(t *testing.T) {
	t.Parallel()

	cacher, _ := timecache.NewTimeCacher(createArgTimeCacher())
	assert.False(t, cacher.IsInterfaceNil())

	providedKey, providedVal := []byte("key"), []byte("val")
	cacher.Put(providedKey, providedVal, len(providedVal))

	v, ok := cacher.Peek(providedKey)
	assert.True(t, ok)
	assert.Equal(t, providedVal, v)

	v, ok = cacher.Peek([]byte("missing key"))
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestTimeCacher_Put(t *testing.T) {
	t.Parallel()

	cacher, _ := timecache.NewTimeCacher(createArgTimeCacher())
	assert.False(t, cacher.IsInterfaceNil())

	numOfPairs := 2
	keys, vals := createKeysVals(numOfPairs)
	evicted := cacher.Put(keys[0], vals[0], len(vals[0]))
	assert.False(t, evicted)
	assert.Equal(t, 1, cacher.Len())
	evicted = cacher.Put(keys[0], vals[1], len(vals[1]))
	assert.False(t, evicted)
	assert.Equal(t, 1, cacher.Len())
}

func TestTimeCacher_Remove(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	cacher, _ := timecache.NewTimeCacher(createArgTimeCacher())
	assert.False(t, cacher.IsInterfaceNil())

	providedKey, providedVal := []byte("key"), []byte("val")
	cacher.Put(providedKey, providedVal, len(providedVal))
	assert.Equal(t, 1, cacher.Len())

	cacher.Remove(nil)
	assert.Equal(t, 1, cacher.Len())

	cacher.Remove(providedKey)
	assert.Equal(t, 0, cacher.Len())

	cacher.Remove(providedKey)
}

func TestTimeCacher_SizeInBytesContained(t *testing.T) {
	t.Parallel()

	cacher, _ := timecache.NewTimeCacher(createArgTimeCacher())
	assert.False(t, cacher.IsInterfaceNil())

	providedKey, providedVal := []byte("key"), []byte("val")
	cacher.Put(providedKey, providedVal, len(providedVal))

	assert.Zero(t, cacher.SizeInBytesContained())
}

func TestTimeCacher_RegisterHandler(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	cacher, _ := timecache.NewTimeCacher(createArgTimeCacher())
	assert.False(t, cacher.IsInterfaceNil())
	cacher.RegisterHandler(func(key []byte, value interface{}) {}, "0")
}

func TestTimeCacher_UnRegisterHandler(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	cacher, _ := timecache.NewTimeCacher(createArgTimeCacher())
	assert.False(t, cacher.IsInterfaceNil())
	cacher.UnRegisterHandler("0")
}

func TestTimeCacher_MaxSize(t *testing.T) {
	t.Parallel()

	cacher, _ := timecache.NewTimeCacher(createArgTimeCacher())
	assert.False(t, cacher.IsInterfaceNil())
	assert.Equal(t, math.MaxInt32, cacher.MaxSize())
}

func TestTimeCacher_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	tc, _ := timecache.NewTimeCacher(createArgTimeCacher())
	numOperations := 1000
	wg := &sync.WaitGroup{}
	wg.Add(numOperations)
	for i := 0; i < numOperations; i++ {
		go func(idx int) {
			time.Sleep(time.Millisecond * 10)

			switch idx % 14 {
			case 0:
				tc.Clear()
			case 1:
				_ = tc.Put(createKeyByteSlice(idx), createValueByteSlice(idx), 0)
			case 2:
				_, _ = tc.Get(createKeyByteSlice(idx))
			case 3:
				_ = tc.Has([]byte(fmt.Sprintf("key%d", idx)))
			case 4:
				_, _ = tc.Peek(createKeyByteSlice(idx))
			case 5:
				_, _ = tc.HasOrAdd(createKeyByteSlice(idx), createValueByteSlice(idx), 0)
			case 6:
				tc.Remove(createKeyByteSlice(idx))
			case 7:
				_ = tc.Keys()
			case 8:
				_ = tc.Len()
			case 9:
				_ = tc.SizeInBytesContained()
			case 10:
				_ = tc.MaxSize()
			case 11:
				tc.RegisterHandler(nil, "")
			case 12:
				tc.UnRegisterHandler("")
			case 13:
				_ = tc.Close()
			default:
				assert.Fail(t, "test setup error, change this line 'switch idx%6{'")
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
}

func createKeyByteSlice(index int) []byte {
	return []byte(fmt.Sprintf("key%d", index))
}

func createValueByteSlice(index int) []byte {
	return []byte(fmt.Sprintf("value%d", index))
}

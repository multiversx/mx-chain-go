package mapTimeCache_test

import (
	"bytes"
	"encoding/gob"
	"sort"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/storage/mapTimeCache"
	"github.com/stretchr/testify/assert"
)

func createArgMapTimeCache() mapTimeCache.ArgMapTimeCacher {
	return mapTimeCache.ArgMapTimeCacher{
		DefaultSpan: time.Minute,
		CacheExpiry: time.Minute,
	}
}

func createKeysVals(noOfPairs int) ([][]byte, [][]byte) {
	keys := make([][]byte, noOfPairs)
	vals := make([][]byte, noOfPairs)
	for i := 0; i < noOfPairs; i++ {
		keys[i] = []byte("k" + string(rune(i)))
		vals[i] = []byte("v" + string(rune(i)))
	}
	return keys, vals
}

func TestNewMapTimeCache(t *testing.T) {
	t.Parallel()

	cacher := mapTimeCache.NewMapTimeCache(createArgMapTimeCache())
	assert.False(t, cacher.IsInterfaceNil())
}

func TestMapTimeCacher_Clear(t *testing.T) {
	t.Parallel()

	cacher := mapTimeCache.NewMapTimeCache(createArgMapTimeCache())
	assert.False(t, cacher.IsInterfaceNil())

	noOfPairs := 3
	providedKeys, providedVals := createKeysVals(noOfPairs)
	for i := 0; i < noOfPairs; i++ {
		evicted := cacher.Put(providedKeys[i], providedVals[i], len(providedVals[i]))
		assert.False(t, evicted)
	}
	assert.Equal(t, noOfPairs, cacher.Len())

	cacher.Clear()
	assert.Equal(t, 0, cacher.Len())
}

func TestMapTimeCacher_Close(t *testing.T) {
	t.Parallel()

	cacher := mapTimeCache.NewMapTimeCache(createArgMapTimeCache())
	assert.False(t, cacher.IsInterfaceNil())

	err := cacher.Close()
	assert.Nil(t, err)
}

func TestMapTimeCacher_Get(t *testing.T) {
	t.Parallel()

	cacher := mapTimeCache.NewMapTimeCache(createArgMapTimeCache())
	assert.False(t, cacher.IsInterfaceNil())

	providedKey, providedVal := []byte("key"), []byte("val")
	evicted := cacher.Put(providedKey, providedVal, len(providedVal))
	assert.False(t, evicted)

	v, ok := cacher.Get(providedKey)
	assert.True(t, ok)
	assert.Equal(t, providedVal, v)

	v, ok = cacher.Get([]byte("missing key"))
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestMapTimeCacher_Has(t *testing.T) {
	t.Parallel()

	cacher := mapTimeCache.NewMapTimeCache(createArgMapTimeCache())
	assert.False(t, cacher.IsInterfaceNil())

	providedKey, providedVal := []byte("key"), []byte("val")
	evicted := cacher.Put(providedKey, providedVal, len(providedVal))
	assert.False(t, evicted)

	assert.True(t, cacher.Has(providedKey))
	assert.False(t, cacher.Has([]byte("missing key")))
}

func TestMapTimeCacher_HasOrAdd(t *testing.T) {
	t.Parallel()

	cacher := mapTimeCache.NewMapTimeCache(createArgMapTimeCache())
	assert.False(t, cacher.IsInterfaceNil())

	providedKey, providedVal := []byte("key"), []byte("val")
	has, added := cacher.HasOrAdd(providedKey, providedVal, len(providedVal))
	assert.False(t, has)
	assert.True(t, added)

	has, added = cacher.HasOrAdd(providedKey, providedVal, len(providedVal))
	assert.True(t, has)
	assert.False(t, added)
}

func TestMapTimeCacher_Keys(t *testing.T) {
	t.Parallel()

	cacher := mapTimeCache.NewMapTimeCache(createArgMapTimeCache())
	assert.False(t, cacher.IsInterfaceNil())

	noOfPairs := 10
	providedKeys, providedVals := createKeysVals(noOfPairs)
	for i := 0; i < noOfPairs; i++ {
		cacher.Put(providedKeys[i], providedVals[i], len(providedVals[i]))
	}

	receivedKeys := cacher.Keys()
	assert.Equal(t, noOfPairs, len(receivedKeys))

	sort.Slice(providedKeys, func(i, j int) bool {
		return bytes.Compare(providedKeys[i], providedKeys[j]) < 0
	})
	sort.Slice(receivedKeys, func(i, j int) bool {
		return bytes.Compare(receivedKeys[i], receivedKeys[j]) < 0
	})

	for i := 0; i < noOfPairs; i++ {
		assert.Equal(t, providedKeys[i], receivedKeys[i])
	}
}

func TestMapTimeCacher_OnSweep(t *testing.T) {
	t.Parallel()

	arg := createArgMapTimeCache()
	arg.CacheExpiry = 2 * time.Second
	arg.DefaultSpan = time.Second
	cacher := mapTimeCache.NewMapTimeCache(arg)
	assert.False(t, cacher.IsInterfaceNil())
	cacher.StartSweeping()

	noOfPairs := 2
	providedKeys, providedVals := createKeysVals(noOfPairs)
	for i := 0; i < noOfPairs; i++ {
		cacher.Put(providedKeys[i], providedVals[i], len(providedVals[i]))
	}
	assert.Equal(t, noOfPairs, cacher.Len())

	time.Sleep(2 * arg.CacheExpiry)
	assert.Equal(t, 0, cacher.Len())
	err := cacher.Close()
	assert.Nil(t, err)
}

func TestMapTimeCacher_Peek(t *testing.T) {
	t.Parallel()

	cacher := mapTimeCache.NewMapTimeCache(createArgMapTimeCache())
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

func TestMapTimeCacher_Put(t *testing.T) {
	t.Parallel()

	cacher := mapTimeCache.NewMapTimeCache(createArgMapTimeCache())
	assert.False(t, cacher.IsInterfaceNil())

	noOfPairs := 2
	keys, vals := createKeysVals(noOfPairs)
	evicted := cacher.Put(keys[0], vals[0], len(vals[0]))
	assert.False(t, evicted)
	assert.Equal(t, 1, cacher.Len())
	evicted = cacher.Put(keys[0], vals[1], len(vals[1]))
	assert.True(t, evicted)
	assert.Equal(t, 1, cacher.Len())
}

func TestMapTimeCacher_Remove(t *testing.T) {
	t.Parallel()

	cacher := mapTimeCache.NewMapTimeCache(createArgMapTimeCache())
	assert.False(t, cacher.IsInterfaceNil())

	providedKey, providedVal := []byte("key"), []byte("val")
	cacher.Put(providedKey, providedVal, len(providedVal))
	assert.Equal(t, 1, cacher.Len())

	cacher.Remove(providedKey)
	assert.Equal(t, 0, cacher.Len())

	cacher.Remove(providedKey)
	if r := recover(); r != nil {
		assert.Fail(t, "should not panic")
	}
}

func TestMapTimeCacher_SizeInBytesContained(t *testing.T) {
	t.Parallel()

	cacher := mapTimeCache.NewMapTimeCache(createArgMapTimeCache())
	assert.False(t, cacher.IsInterfaceNil())

	providedKey, providedVal := []byte("key"), []byte("val")
	cacher.Put(providedKey, providedVal, len(providedVal))

	b := new(bytes.Buffer)
	err := gob.NewEncoder(b).Encode(providedVal)
	assert.Nil(t, err)
	assert.Equal(t, uint64(b.Len()), cacher.SizeInBytesContained())
}

func TestMapTimeCacher_RegisterHandler(t *testing.T) {
	t.Parallel()

	cacher := mapTimeCache.NewMapTimeCache(createArgMapTimeCache())
	assert.False(t, cacher.IsInterfaceNil())
	cacher.RegisterHandler(func(key []byte, value interface{}) {}, "0")

	if r := recover(); r != nil {
		assert.Fail(t, "should not panic")
	}
}

func TestMapTimeCacher_UnRegisterHandler(t *testing.T) {
	t.Parallel()

	cacher := mapTimeCache.NewMapTimeCache(createArgMapTimeCache())
	assert.False(t, cacher.IsInterfaceNil())
	cacher.UnRegisterHandler("0")

	if r := recover(); r != nil {
		assert.Fail(t, "should not panic")
	}
}

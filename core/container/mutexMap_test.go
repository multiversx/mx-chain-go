package container

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMutexMap(t *testing.T) {
	t.Parallel()

	mm := NewMutexMap()

	assert.NotNil(t, mm)
}

//------- Get

func TestMutexMap_GetAndSet(t *testing.T) {
	t.Parallel()

	mm := NewMutexMap()

	val, ok := mm.Get("inexistent key")
	assert.Nil(t, val)
	assert.False(t, ok)

	key := "existent key"
	valToStore := 56
	mm.Set(key, valToStore)

	val, ok = mm.Get(key)
	assert.Equal(t, valToStore, val)
	assert.True(t, ok)
}

func TestMutexMap_GetConcurrent(t *testing.T) {
	t.Parallel()

	mm := NewMutexMap()
	key := "existent key"
	valToStore := 56
	mm.Set(key, valToStore)

	numIterations := 100
	wg := &sync.WaitGroup{}
	wg.Add(numIterations)
	for i := 0; i < numIterations; i++ {
		go func() {
			val, ok := mm.Get(key)
			assert.Equal(t, valToStore, val)
			assert.True(t, ok)

			wg.Done()
		}()
	}

	wg.Wait()
}

//------- Set

func TestMutexMap_SetConcurrent(t *testing.T) {
	t.Parallel()

	mm := NewMutexMap()
	key := "existent key"
	valToStore := 56

	numIterations := 100
	wg := &sync.WaitGroup{}
	wg.Add(numIterations)
	for i := 0; i < numIterations; i++ {
		go func() {
			mm.Set(key, valToStore)
			wg.Done()
		}()
	}

	wg.Wait()

	val, ok := mm.Get(key)
	assert.Equal(t, valToStore, val)
	assert.True(t, ok)
}

//------- Insert

func TestMutexMap_Insert(t *testing.T) {
	t.Parallel()

	mm := NewMutexMap()

	key := "key"
	val1 := 56
	val2 := 78

	//operation succeed
	ok := mm.Insert(key, val1)
	assert.True(t, ok)

	//test inner value
	recoveredValue, _ := mm.Get(key)
	assert.Equal(t, val1, recoveredValue)

	//rewrite will fail
	ok = mm.Insert(key, val2)
	assert.False(t, ok)

	//test inner value
	recoveredValue, _ = mm.Get(key)
	assert.Equal(t, val1, recoveredValue)
}

func TestMutexMap_InsertConcurrent(t *testing.T) {
	t.Parallel()

	mm := NewMutexMap()

	numIterations := 100
	wg := &sync.WaitGroup{}
	wg.Add(numIterations)
	for i := 0; i < numIterations; i++ {
		go func(idx int) {
			key := fmt.Sprintf("%d", idx)
			ok := mm.Insert(key, struct{}{})
			assert.True(t, ok)

			wg.Done()
		}(i)
	}

	wg.Wait()
}

//------- Remove

func TestMutexMap_Remove(t *testing.T) {
	t.Parallel()

	mm := NewMutexMap()

	key := "key"
	mm.Remove(key)

	mm.Set(key, struct{}{})
	valRecovered, _ := mm.Get(key)
	assert.NotNil(t, valRecovered)

	mm.Remove(key)

	valRecovered, _ = mm.Get(key)
	assert.Nil(t, valRecovered)
}

func TestMutexMap_RemoveConcurrentShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should not have paniced %v", r))
		}
	}()

	mm := NewMutexMap()

	numIterations := 100
	wg := &sync.WaitGroup{}
	wg.Add(numIterations)
	for i := 0; i < numIterations; i++ {
		go func(idx int) {
			mm.Remove("key")

			wg.Done()
		}(i)
	}

	wg.Wait()
}

//------- Keys

func TestMutexMap_Keys(t *testing.T) {
	t.Parallel()

	mm := NewMutexMap()

	key1 := "key1"
	key2 := "key2"
	key3 := "key3"

	keys := mm.Keys()
	assert.Equal(t, 0, len(keys))

	mm.Set(key1, struct{}{})
	mm.Set(key2, struct{}{})
	mm.Set(key3, struct{}{})

	keys = mm.Keys()
	require.Equal(t, 3, len(keys))
	stringKeys := make([]string, len(keys))
	for idx, k := range keys {
		stringKeys[idx] = k.(string)
	}

	sort.Slice(stringKeys, func(i, j int) bool {
		return strings.Compare(stringKeys[i], stringKeys[j]) < 0
	})

	assert.Equal(t, key1, stringKeys[0])
	assert.Equal(t, key2, stringKeys[1])
	assert.Equal(t, key3, stringKeys[2])
}

func TestMutexMap_KeysConcurrent(t *testing.T) {
	t.Parallel()

	mm := NewMutexMap()
	mm.Set("key1", struct{}{})
	mm.Set("key2", struct{}{})
	mm.Set("key3", struct{}{})

	numIterations := 100
	wg := &sync.WaitGroup{}
	wg.Add(numIterations)
	for i := 0; i < numIterations; i++ {
		go func(idx int) {
			keys := mm.Keys()
			assert.Equal(t, 3, len(keys))

			wg.Done()
		}(i)
	}

	wg.Wait()
}

//------- Values

func TestMutexMap_Values(t *testing.T) {
	t.Parallel()

	mm := NewMutexMap()
	values := mm.Values()
	assert.Equal(t, 0, len(values))

	mm.Set("1", "value1")
	mm.Set("2", "value2")
	mm.Set("3", "value3")

	values = mm.Values()
	require.Equal(t, 3, len(values))
	require.ElementsMatch(t, []interface{}{"value1", "value2", "value3"}, values)
}

func TestMutexMap_ValuesConcurrent(t *testing.T) {
	t.Parallel()

	mm := NewMutexMap()

	numIterations := 100
	wg := &sync.WaitGroup{}
	wg.Add(numIterations)
	for i := 0; i < numIterations; i++ {
		go func(i int) {
			mm.Set(i, i)
			assert.GreaterOrEqual(t, len(mm.Values()), 1)
			wg.Done()
		}(i)
	}

	wg.Wait()
	require.Equal(t, numIterations, len(mm.Values()))
}

//------- Len

func TestMutexMap_Len(t *testing.T) {
	t.Parallel()

	mm := NewMutexMap()

	key1 := "key1"
	key2 := "key2"
	key3 := "key3"

	assert.Equal(t, 0, mm.Len())

	mm.Set(key1, struct{}{})
	assert.Equal(t, 1, mm.Len())

	mm.Set(key2, struct{}{})
	assert.Equal(t, 2, mm.Len())

	mm.Set(key3, struct{}{})
	assert.Equal(t, 3, mm.Len())
}

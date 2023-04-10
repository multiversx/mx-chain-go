package trieValuesCache

import (
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieValuesCache(t *testing.T) {
	t.Parallel()

	t.Run("invalid maxNumEntries", func(t *testing.T) {
		t.Parallel()

		tvc, err := NewTrieValuesCache(0)
		assert.True(t, check.IfNil(tvc))
		assert.NotNil(t, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tvc, err := NewTrieValuesCache(10)
		assert.False(t, check.IfNil(tvc))
		assert.Nil(t, err)
	})
}

func TestTrieValuesCache_Put(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tvc, _ := NewTrieValuesCache(10)
		key := []byte("key")
		trieVal := []byte("trie value")
		expectedVal := core.TrieData{
			Key:     key,
			Value:   trieVal,
			Version: 1,
		}

		tvc.Put(key, expectedVal)
		assert.Equal(t, 1, tvc.numEntries)
		retrievedVal, ok := tvc.cache[string(key)]
		assert.True(t, ok)
		assert.Equal(t, expectedVal, retrievedVal)
	})

	t.Run("full cache", func(t *testing.T) {
		t.Parallel()

		maxNumEntries := 10
		tvc, _ := NewTrieValuesCache(maxNumEntries)
		for i := 0; i < maxNumEntries; i++ {
			tvc.Put([]byte("key"+strconv.Itoa(i)), core.TrieData{})
		}
		assert.Equal(t, 10, tvc.numEntries)

		extraKey := []byte("extra key")
		tvc.Put(extraKey, core.TrieData{})

		assert.Equal(t, 10, tvc.numEntries)
		_, ok := tvc.cache[string(extraKey)]
		assert.False(t, ok)
	})
}

func TestTrieValuesCache_Get(t *testing.T) {
	t.Parallel()

	t.Run("empty cache", func(t *testing.T) {
		t.Parallel()

		expectedVal := core.TrieData{}
		tvc, _ := NewTrieValuesCache(10)
		trieVal, ok := tvc.Get([]byte("key"))
		assert.False(t, ok)
		assert.Equal(t, expectedVal.Key, trieVal.Key)
		assert.Equal(t, expectedVal.Value, trieVal.Value)
		assert.Equal(t, expectedVal.Version, trieVal.Version)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tvc, _ := NewTrieValuesCache(10)
		key := []byte("key")
		trieVal := []byte("trie value")
		expectedVal := core.TrieData{
			Key:     key,
			Value:   trieVal,
			Version: 1,
		}

		tvc.cache[string(key)] = expectedVal
		tvc.numEntries = 1

		retrievedVal, ok := tvc.Get(key)
		assert.True(t, ok)
		assert.Equal(t, expectedVal, retrievedVal)

		expectedVal = core.TrieData{}
		val, ok := tvc.Get([]byte("non existing key"))
		assert.False(t, ok)
		assert.Equal(t, expectedVal.Key, val.Key)
		assert.Equal(t, expectedVal.Value, val.Value)
		assert.Equal(t, expectedVal.Version, val.Version)
	})
}

func TestTrieValuesCache_Clean(t *testing.T) {
	t.Parallel()

	tvc, _ := NewTrieValuesCache(10)
	key := []byte("key")
	tvc.Put(key, core.TrieData{})
	assert.Equal(t, 1, tvc.numEntries)

	tvc.Clean()
	assert.Equal(t, 0, tvc.numEntries)
	_, ok := tvc.Get(key)
	assert.False(t, ok)
}

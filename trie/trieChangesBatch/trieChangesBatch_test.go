package trieChangesBatch

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieChangesBatch(t *testing.T) {
	t.Parallel()

	tcb := NewTrieChangesBatch()
	assert.False(t, check.IfNil(tcb))
	assert.Equal(t, 0, len(tcb.insertedData))
	assert.Equal(t, 0, len(tcb.deletedKeys))
}

func TestTrieChangesBatch_Add(t *testing.T) {
	t.Parallel()

	keyForInsertion := []byte("keyForInsertion")
	dataForInsertion := core.TrieData{
		Key:     []byte("trieKey"),
		Value:   []byte("trieValue"),
		Version: core.NotSpecified,
	}

	tcb := NewTrieChangesBatch()
	tcb.deletedKeys[string(keyForInsertion)] = struct{}{}

	tcb.Add(keyForInsertion, dataForInsertion)
	assert.Equal(t, 0, len(tcb.deletedKeys))
	assert.Equal(t, 1, len(tcb.insertedData))
	assert.Equal(t, dataForInsertion, tcb.insertedData[string(keyForInsertion)])
}

func TestTrieChangesBatch_MarkForRemoval(t *testing.T) {
	t.Parallel()

	keyForDeletion := []byte("keyForDeletion")

	tcb := NewTrieChangesBatch()
	tcb.insertedData[string(keyForDeletion)] = core.TrieData{
		Key:     []byte("trieKey"),
		Value:   []byte("trieValue"),
		Version: core.NotSpecified,
	}

	tcb.MarkForRemoval(keyForDeletion)
	assert.Equal(t, 0, len(tcb.insertedData))
	assert.Equal(t, 1, len(tcb.deletedKeys))
	_, found := tcb.deletedKeys[string(keyForDeletion)]
	assert.True(t, found)
}

func TestTrieChangesBatch_Get(t *testing.T) {
	t.Parallel()

	t.Run("key exists in insertedData", func(t *testing.T) {
		t.Parallel()

		key := []byte("key")
		value := []byte("value")

		tcb := NewTrieChangesBatch()
		tcb.insertedData[string(key)] = core.TrieData{
			Key:     key,
			Value:   value,
			Version: core.NotSpecified,
		}

		data, foundInBatch := tcb.Get(key)
		assert.True(t, foundInBatch)
		assert.Equal(t, value, data)
	})
	t.Run("key exists in deletedKeys", func(t *testing.T) {
		t.Parallel()

		key := []byte("key")
		tcb := NewTrieChangesBatch()
		tcb.deletedKeys[string(key)] = struct{}{}

		data, foundInBatch := tcb.Get(key)
		assert.True(t, foundInBatch)
		assert.Nil(t, data)
	})
	t.Run("key does not exist in batch", func(t *testing.T) {
		t.Parallel()

		key := []byte("key")
		tcb := NewTrieChangesBatch()

		data, foundInBatch := tcb.Get(key)
		assert.False(t, foundInBatch)
		assert.Nil(t, data)
	})
}

func TestTrieChangesBatch_GetSortedDataForInsertion(t *testing.T) {
	t.Parallel()

	tcb := NewTrieChangesBatch()

	tcb.insertedData["key3"] = core.TrieData{}
	tcb.insertedData["key1"] = core.TrieData{}
	tcb.insertedData["key2"] = core.TrieData{}

	keys, data := tcb.GetSortedDataForInsertion()
	assert.Equal(t, []string{"key1", "key2", "key3"}, keys)
	assert.Equal(t, 3, len(data))
}

func TestTrieChangesBatch_GetSortedDataForRemoval(t *testing.T) {
	t.Parallel()

	tcb := NewTrieChangesBatch()

	tcb.deletedKeys["key3"] = struct{}{}
	tcb.deletedKeys["key1"] = struct{}{}
	tcb.deletedKeys["key2"] = struct{}{}

	keys := tcb.GetSortedDataForRemoval()
	assert.Equal(t, []string{"key1", "key2", "key3"}, keys)
}

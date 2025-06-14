package trieChangesBatch

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieChangesBatch(t *testing.T) {
	t.Parallel()

	tcb := NewTrieChangesBatch("")
	assert.False(t, check.IfNil(tcb))
	assert.Equal(t, 0, len(tcb.insertedData))
	assert.Equal(t, 0, len(tcb.deletedKeys))
}

func TestTrieChangesBatch_Add(t *testing.T) {
	t.Parallel()

	dataForInsertion := core.TrieData{
		Key:     []byte("trieKey"),
		Value:   []byte("trieValue"),
		Version: core.NotSpecified,
	}

	tcb := NewTrieChangesBatch("")
	tcb.deletedKeys[string(dataForInsertion.Key)] = core.TrieData{Key: dataForInsertion.Key}

	tcb.Add(dataForInsertion)
	assert.Equal(t, 0, len(tcb.deletedKeys))
	assert.Equal(t, 1, len(tcb.insertedData))
	assert.Equal(t, dataForInsertion, tcb.insertedData[string(dataForInsertion.Key)])
}

func TestTrieChangesBatch_MarkForRemoval(t *testing.T) {
	t.Parallel()

	keyForDeletion := []byte("keyForDeletion")

	tcb := NewTrieChangesBatch("")
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

		tcb := NewTrieChangesBatch("")
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
		tcb := NewTrieChangesBatch("")
		tcb.deletedKeys[string(key)] = core.TrieData{Key: key}

		data, foundInBatch := tcb.Get(key)
		assert.True(t, foundInBatch)
		assert.Nil(t, data)
	})
	t.Run("key does not exist in batch", func(t *testing.T) {
		t.Parallel()

		key := []byte("key")
		tcb := NewTrieChangesBatch("")

		data, foundInBatch := tcb.Get(key)
		assert.False(t, foundInBatch)
		assert.Nil(t, data)
	})
}

func TestTrieChangesBatch_GetSortedDataForInsertion(t *testing.T) {
	t.Parallel()

	tcb := NewTrieChangesBatch("")
	tcb.insertedData["key3"] = core.TrieData{Key: []byte("key3")}
	tcb.insertedData["key1"] = core.TrieData{Key: []byte("key1")}
	tcb.insertedData["key2"] = core.TrieData{Key: []byte("key2")}

	data := tcb.GetSortedDataForInsertion()
	assert.Equal(t, 3, len(data))
	assert.Equal(t, keyBuilder.KeyBytesToHex([]byte("key1")), data[0].Key)
	assert.Equal(t, keyBuilder.KeyBytesToHex([]byte("key2")), data[1].Key)
	assert.Equal(t, keyBuilder.KeyBytesToHex([]byte("key3")), data[2].Key)
}

func TestTrieChangesBatch_GetSortedDataForRemoval(t *testing.T) {
	t.Parallel()

	tcb := NewTrieChangesBatch("")

	key1 := "key1"
	key2 := "key2"
	key3 := "key3"

	tcb.deletedKeys[key3] = core.TrieData{Key: []byte(key3)}
	tcb.deletedKeys[key1] = core.TrieData{Key: []byte(key1)}
	tcb.deletedKeys[key2] = core.TrieData{Key: []byte(key2)}

	data := tcb.GetSortedDataForRemoval()
	assert.Equal(t, 3, len(data))
	assert.Equal(t, keyBuilder.KeyBytesToHex([]byte(key1)), data[0].Key)
	assert.Equal(t, keyBuilder.KeyBytesToHex([]byte(key2)), data[1].Key)
	assert.Equal(t, keyBuilder.KeyBytesToHex([]byte(key3)), data[2].Key)
}

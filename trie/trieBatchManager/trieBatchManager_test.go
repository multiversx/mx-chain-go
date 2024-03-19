package trieBatchManager

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieBatchManager(t *testing.T) {
	t.Parallel()

	tbm := NewTrieBatchManager()
	assert.False(t, check.IfNil(tbm))

	assert.False(t, tbm.isUpdateInProgress)
	assert.False(t, check.IfNil(tbm.currentBatch))
	assert.True(t, check.IfNil(tbm.tempBatch))
}

func TestTrieBatchManager_TrieUpdateInProgress(t *testing.T) {
	t.Parallel()

	tbm := NewTrieBatchManager()
	tbm.currentBatch.Add([]byte("key1"), core.TrieData{})
	tbm.currentBatch.Add([]byte("key2"), core.TrieData{})
	assert.False(t, tbm.isUpdateInProgress)
	assert.True(t, check.IfNil(tbm.tempBatch))

	batchManager, err := tbm.MarkTrieUpdateInProgress()
	assert.True(t, tbm.isUpdateInProgress)
	assert.False(t, check.IfNil(tbm.tempBatch))
	assert.Nil(t, err)
	_, found := batchManager.Get([]byte("key1"))
	assert.True(t, found)
	_, found = batchManager.Get([]byte("key2"))
	assert.True(t, found)

	batchManager, err = tbm.MarkTrieUpdateInProgress()
	assert.True(t, check.IfNil(batchManager))
	assert.Equal(t, ErrTrieUpdateInProgress, err)

	tbm.MarkTrieUpdateCompleted()
	assert.False(t, tbm.isUpdateInProgress)
	assert.True(t, check.IfNil(tbm.tempBatch))
}

func TestTrieBatchManager_AddUpdatesCurrentBatch(t *testing.T) {
	t.Parallel()

	tbm := NewTrieBatchManager()
	tbm.Add([]byte("key1"), core.TrieData{})
	_, found := tbm.currentBatch.Get([]byte("key1"))
	assert.True(t, found)

	_, _ = tbm.MarkTrieUpdateInProgress()
	tbm.Add([]byte("key2"), core.TrieData{})
	_, found = tbm.currentBatch.Get([]byte("key2"))
	assert.True(t, found)
}

func TestTrieBatchManager_Get(t *testing.T) {
	t.Parallel()

	key := []byte("key")
	value := []byte("value")

	t.Run("key exists in currentBatch", func(t *testing.T) {
		t.Parallel()

		tbm := NewTrieBatchManager()
		tbm.currentBatch.Add(key, core.TrieData{
			Value: value,
		})

		data, found := tbm.Get(key)
		assert.True(t, found)
		assert.Equal(t, value, data)
	})
	t.Run("check temp batch only if update in progress", func(t *testing.T) {
		t.Parallel()

		tbm := NewTrieBatchManager()
		tbm.currentBatch.Add(key, core.TrieData{
			Value: value,
		})
		_, _ = tbm.MarkTrieUpdateInProgress()
		_, found := tbm.currentBatch.Get(key)
		assert.False(t, found)

		data, found := tbm.Get(key)
		assert.True(t, found)
		assert.Equal(t, value, data)
	})
	t.Run("code does not panic if temp batch is nil", func(t *testing.T) {
		t.Parallel()

		tbm := NewTrieBatchManager()
		tbm.isUpdateInProgress = true
		tbm.tempBatch = nil

		data, found := tbm.Get(key)
		assert.False(t, found)
		assert.Nil(t, data)
	})
	t.Run("key does not exist", func(t *testing.T) {
		t.Parallel()

		tbm := NewTrieBatchManager()

		data, found := tbm.Get(key)
		assert.False(t, found)
		assert.Nil(t, data)
	})
}

func TestTrieBatchManager_MarkForRemovalUpdatesCurrentBatch(t *testing.T) {
	t.Parallel()

	key := []byte("key")
	tbm := NewTrieBatchManager()
	_, _ = tbm.MarkTrieUpdateInProgress()
	tbm.MarkForRemoval(key)

	_, found := tbm.currentBatch.Get(key)
	assert.True(t, found)

	_, found = tbm.tempBatch.Get(key)
	assert.False(t, found)
}

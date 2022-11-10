package disabled

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewDisabledTrackableDataTrie(t *testing.T) {
	t.Parallel()

	t.Run("new disabledTrackableDataTrie", func(t *testing.T) {
		t.Parallel()

		assert.False(t, check.IfNil(NewDisabledTrackableDataTrie()))
	})

	t.Run("retrieve value", func(t *testing.T) {
		t.Parallel()

		dtdt := NewDisabledTrackableDataTrie()

		val, depth, err := dtdt.RetrieveValue(nil)
		assert.Nil(t, err)
		assert.Equal(t, uint32(0), depth)
		assert.Equal(t, 0, len(val))
	})

	t.Run("saveKeyValue", func(t *testing.T) {
		t.Parallel()

		dtdt := NewDisabledTrackableDataTrie()

		err := dtdt.SaveKeyValue(nil, nil)
		assert.Nil(t, err)
	})

	t.Run("set and get data trie", func(t *testing.T) {
		t.Parallel()

		dtdt := NewDisabledTrackableDataTrie()
		isDisabledDataTrieHandler := false
		dtdt.SetDataTrie(nil)
		tr := dtdt.DataTrie()

		switch tr.(type) {
		case *disabledDataTrieHandler:
			isDisabledDataTrieHandler = true
		default:
			assert.Fail(t, "this should not have been called")
		}
		assert.True(t, isDisabledDataTrieHandler)
	})

	t.Run("save dirty data", func(t *testing.T) {
		t.Parallel()

		dtdt := NewDisabledTrackableDataTrie()

		oldValues, err := dtdt.SaveDirtyData(nil)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(oldValues))
	})
}

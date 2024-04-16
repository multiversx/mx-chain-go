package trieNodeData

import (
	"testing"

	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/stretchr/testify/assert"
)

func TestBaseNodeData(t *testing.T) {
	t.Parallel()

	t.Run(" empty base node data", func(t *testing.T) {
		t.Parallel()

		bnd := &baseNodeData{}
		assert.Nil(t, bnd.GetData())
		assert.Nil(t, bnd.GetKeyBuilder())
		assert.Equal(t, uint64(0), bnd.Size())
	})
	t.Run(" base node data with data", func(t *testing.T) {
		t.Parallel()

		data := []byte("data")
		key := []byte("key")
		kb := keyBuilder.NewKeyBuilder()
		kb.BuildKey(key)
		bnd := &baseNodeData{
			data:       data,
			keyBuilder: kb,
		}

		assert.Equal(t, data, bnd.GetData())
		assert.Equal(t, kb, bnd.GetKeyBuilder())
		assert.Equal(t, uint64(len(data)+len(key)), bnd.Size())
	})
}

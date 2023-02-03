package parsers

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewMainTrieLeafParser(t *testing.T) {
	t.Parallel()

	t.Run("new mainTrieLeafParser", func(t *testing.T) {
		t.Parallel()

		assert.False(t, check.IfNil(NewMainTrieLeafParser()))
	})

	t.Run("parse leaf", func(t *testing.T) {
		t.Parallel()

		key := []byte("key")
		value := []byte("value")
		dtlp := NewMainTrieLeafParser()

		keyValHolder, err := dtlp.ParseLeaf(key, value, core.NotSpecified)
		assert.Nil(t, err)
		assert.Equal(t, key, keyValHolder.Key())
		assert.Equal(t, value, keyValHolder.Value())
	})
}

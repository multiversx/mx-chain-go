package parsers

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieLeafParserV1(t *testing.T) {
	t.Parallel()

	t.Run("new trieLeafParserV1", func(t *testing.T) {
		t.Parallel()

		assert.False(t, check.IfNil(NewTrieLeafParserV1()))
	})

	t.Run("parse leaf", func(t *testing.T) {
		t.Parallel()

		key := []byte("key")
		value := []byte("value")
		dtlp := NewTrieLeafParserV1()

		keyValHolder, err := dtlp.ParseLeaf(key, value)
		assert.Nil(t, err)
		assert.Equal(t, key, keyValHolder.Key())
		assert.Equal(t, value, keyValHolder.Value())
	})
}

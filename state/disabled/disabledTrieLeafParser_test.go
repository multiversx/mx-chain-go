package disabled

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewDisabledTrieLeafParser(t *testing.T) {
	t.Parallel()

	t.Run("new disabledTrieLeafParser", func(t *testing.T) {
		t.Parallel()

		assert.False(t, check.IfNil(NewDisabledTrieLeafParser()))
	})

	t.Run("parse leaf", func(t *testing.T) {
		t.Parallel()

		key := []byte("key")
		value := []byte("value")
		dtlp := NewDisabledTrieLeafParser()

		keyValHolder, err := dtlp.ParseLeaf(key, value)
		assert.Nil(t, err)
		assert.Equal(t, key, keyValHolder.Key())
		assert.Equal(t, value, keyValHolder.Value())
	})
}

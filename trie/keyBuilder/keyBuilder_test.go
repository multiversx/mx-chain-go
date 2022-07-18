package keyBuilder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyBuilder_Clone(t *testing.T) {
	t.Parallel()

	kb := NewKeyBuilder()
	kb.BuildKey([]byte("dog"))

	clonedKb := kb.Clone()
	clonedKb.BuildKey([]byte("e"))

	originalKey, _ := kb.GetKey()
	modifiedKey, _ := clonedKb.GetKey()
	assert.NotEqual(t, originalKey, modifiedKey)
}

func TestHexToTrieKeyBytes(t *testing.T) {
	t.Parallel()

	reversedHexDoeKey := []byte{5, 6, 15, 6, 4, 6, 16}
	reversedHexDogKey := []byte{7, 6, 15, 6, 4, 6, 16}

	var test = []struct {
		key, hex []byte
	}{
		{reversedHexDoeKey, []byte("doe")},
		{reversedHexDogKey, []byte("dog")},
	}

	for i := range test {
		key, err := hexToTrieKeyBytes(test[i].key)
		assert.Nil(t, err)
		assert.Equal(t, test[i].hex, key)
	}
}

func TestHexToTrieKeyBytesInvalidLength(t *testing.T) {
	t.Parallel()

	key, err := hexToTrieKeyBytes([]byte{6, 4, 6, 15, 6, 5})
	assert.Nil(t, key)
	assert.Equal(t, ErrInvalidLength, err)
}

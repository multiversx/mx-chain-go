package trie2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyBytesToHex(t *testing.T) {
	t.Parallel()
	var test = []struct {
		key, hex []byte
	}{
		{[]byte("doe"), []byte{6, 4, 6, 15, 6, 5, 16}},
		{[]byte("dog"), []byte{6, 4, 6, 15, 6, 7, 16}},
	}

	for i := range test {
		assert.Equal(t, test[i].hex, keyBytesToHex(test[i].key))
	}

}

func TestPrefixLen(t *testing.T) {
	t.Parallel()
	var test = []struct {
		a, b   []byte
		length int
	}{
		{[]byte("doe"), []byte("dog"), 2},
		{[]byte("dog"), []byte("dogglesworth"), 3},
		{[]byte("mouse"), []byte("mouse"), 5},
		{[]byte("caterpillar"), []byte("cats"), 3},
		{[]byte("caterpillar"), []byte(""), 0},
		{[]byte(""), []byte("caterpillar"), 0},
		{[]byte("a"), []byte("caterpillar"), 0},
	}

	for i := range test {
		assert.Equal(t, test[i].length, prefixLen(test[i].a, test[i].b))
	}
}

func TestHasTerm(t *testing.T) {
	t.Parallel()
	assert.True(t, hasTerm([]byte{6, 4, 6, 3, 2, 1, 16}))
	assert.False(t, hasTerm([]byte{6, 4, 6, 3, 2, 1}))
}

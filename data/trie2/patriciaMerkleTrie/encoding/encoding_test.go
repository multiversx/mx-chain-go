package encoding_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2/patriciaMerkleTrie/encoding"

	"github.com/stretchr/testify/assert"
)

func TestKeyBytesToHex(t *testing.T) {
	var test = []struct {
		key, hex []byte
	}{
		{[]byte("doe"), []byte{6, 4, 6, 15, 6, 5, 16}},
		{[]byte("dog"), []byte{6, 4, 6, 15, 6, 7, 16}},
	}

	for i := range test {
		assert.Equal(t, test[i].hex, encoding.KeyBytesToHex(test[i].key))
	}

}

func TestPrefixLen(t *testing.T) {
	var test = []struct {
		a, b   []byte
		length int
	}{
		{[]byte("doe"), []byte("dog"), 2},
		{[]byte("dog"), []byte("dogglesworth"), 3},
		{[]byte("mouse"), []byte("mouse"), 5},
		{[]byte("caterpillar"), []byte("cats"), 3},
	}

	for i := range test {
		assert.Equal(t, test[i].length, encoding.PrefixLen(test[i].a, test[i].b))
	}
}

func TestHasPrefix(t *testing.T) {
	val1 := encoding.KeyBytesToHex([]byte("dog"))
	val2 := []byte{6, 4, 6, 15, 6, 7}
	val3 := encoding.KeyBytesToHex([]byte("doe"))

	assert.True(t, encoding.HasPrefix(val1, val2))
	assert.False(t, encoding.HasPrefix(val3, val2))
}

func TestHasTerm(t *testing.T) {
	key := encoding.KeyBytesToHex([]byte("dodge"))

	assert.True(t, encoding.HasTerm(key))
}
func TestHexToKeyBytes(t *testing.T) {
	var test = []struct {
		key, hex []byte
	}{
		{[]byte("doe"), []byte{6, 4, 6, 15, 6, 5, 16}},
		{[]byte("dog"), []byte{6, 4, 6, 15, 6, 7, 16}},
	}

	for i := range test {
		assert.Equal(t, test[i].key, encoding.HexToKeyBytes(test[i].hex))
	}
}

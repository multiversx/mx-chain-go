package encoding_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2/encoding"

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

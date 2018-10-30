package bloom_test

import (
	"hash"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/storage/bloom"
	"github.com/dchest/blake2b"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/sha3"
)

func TestNewFilter(t *testing.T) {
	_, err := bloom.NewFilter(200, []func() hash.Hash{sha3.NewLegacyKeccak256, blake2b.New256})

	assert.Nil(t, err, "Error creating new bloom filter")
}

func TestNewFilterWithSmallSize(t *testing.T) {
	_, err := bloom.NewFilter(1, []func() hash.Hash{sha3.NewLegacyKeccak256, blake2b.New256})

	assert.NotNil(t, err, "Expected nil")
}

func TestNewDefaultFilter(t *testing.T) {
	_, err := bloom.NewDefaultFilter()

	assert.Nil(t, err, "Error creating new default bloom filter")
}

func TestFilter(t *testing.T) {
	b, _ := bloom.NewDefaultFilter()

	var testTable = []struct {
		in       []byte
		expected bool
	}{
		{[]byte("12345"), true},
		{[]byte(" "), true},
		{[]byte("BloomFilter"), true},
		{[]byte("test"), true},
	}

	for _, val := range testTable {
		b.Add(val.in)
		assert.Equal(t, val.expected, b.Test(val.in), "Expected value to be there")
	}

	b.Clear()

	for _, val := range testTable {
		assert.Equal(t, false, b.Test(val.in), "Expected bloom filter to be empty")
	}

}

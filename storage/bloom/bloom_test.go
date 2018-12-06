package bloom_test

import (
	"strconv"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/fnv"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/bloom"

	"github.com/stretchr/testify/assert"
)

func TestNewFilter(t *testing.T) {
	_, err := bloom.NewFilter(200, []hashing.Hasher{keccak.Keccak{}, blake2b.Blake2b{}, fnv.Fnv{}})

	assert.Nil(t, err, "Error creating new bloom filter")
}

func TestNewFilterWithSmallSize(t *testing.T) {
	_, err := bloom.NewFilter(1, []hashing.Hasher{keccak.Keccak{}, blake2b.Blake2b{}})

	assert.NotNil(t, err, "Expected nil")
}

func TestNewFilterWithZeroHashFunctions(t *testing.T) {
	_, err := bloom.NewFilter(2048, []hashing.Hasher{})

	assert.NotNil(t, err, "Expected nil")
}

func TestFilter(t *testing.T) {
	b := bloom.NewDefaultFilter()

	var testTable = []struct {
		in       []byte
		expected bool
	}{
		{[]byte("12345"), true},
		{[]byte(" "), true},
		{[]byte("BloomFilter"), true},
		{[]byte("test"), true},
		{[]byte("i3419"), true},
		{[]byte("j6147"), true},
	}

	for _, val := range testTable {
		b.Add(val.in)
		assert.Equal(t, val.expected, b.MayContain(val.in), "Expected value to be there")
	}

	b.Clear()

	for _, val := range testTable {
		assert.Equal(t, false, b.MayContain(val.in), "Expected bloom filter to be empty")
	}

}

func TestConcurrency(t *testing.T) {
	b := bloom.NewDefaultFilter()

	wg := sync.WaitGroup{}
	wg.Add(2)

	maxIterations := 10000

	addValues := func(base string) {
		for i := 0; i < maxIterations; i++ {
			b.Add([]byte(base + strconv.Itoa(i)))
		}

		wg.Done()
	}

	go addValues("i")
	go addValues("j")

	wg.Wait()

	for i := 0; i < maxIterations; i++ {
		assert.True(t, b.MayContain([]byte("i"+strconv.Itoa(i))), "i"+strconv.Itoa(i))
		assert.True(t, b.MayContain([]byte("j"+strconv.Itoa(i))), "j"+strconv.Itoa(i))
	}
}

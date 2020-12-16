// A Bloom filter is a probabilistic data structures that is used to test whether an element is a member of a set.
// Using a bloom filter, you know that an element "possibly is in set" or "definitely not in set".
// Initially, all of the n bits from the bloom filter are set to 0. When an element is added, we hash that element with
// several hashing functions (you can choose how many and which ones). Every hash represents an index pointing to a
// position in the bloom filter. The bits that are indicated by the indexes are set to 1. When you want to check if an
// element is in the bloom filter, you hash the element with the same hashing functions, and it will return the same indexes.
// If all the bits pointed by the indexes are set to 1, it means that the element might have been added, otherwise
// (if one bit is set to 0), then the element was definitely not added to the filter.
//
// https://en.wikipedia.org/wiki/Bloom_filter

package bloom

import (
	"encoding/binary"
	"errors"

	"sync"

	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/hashing/fnv"
	"github.com/ElrondNetwork/elrond-go/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ storage.BloomFilter = (*Bloom)(nil)

const (
	bitsInByte = 8
)

// Bloom represents a bloom filter. It holds the filter itself, the hashing functions that must be
// applied to values that are added to the filter and a mutex to handle concurrent accesses to the filter
type Bloom struct {
	filter   []byte
	hashFunc []hashing.Hasher
	mutex    sync.Mutex
}

// NewFilter returns a new Bloom object with the given size and
// hashing functions implementation. It returns an error if there are no hashing
// functions, or if the size of the filter is too small
func NewFilter(size uint, h []hashing.Hasher) (*Bloom, error) {

	if size <= uint(len(h)) {
		return nil, errors.New("filter size is too low")
	}

	if len(h) == 0 {
		return nil, errors.New("too few hashing functions")
	}

	return &Bloom{
		filter:   make([]byte, size),
		hashFunc: h,
	}, nil
}

// NewDefaultFilter returns a new Bloom object with a filter size of 2048 bytes
// and implementations of blake2b, sha3-keccak and fnv128a hashing functions
func NewDefaultFilter() *Bloom {
	return &Bloom{
		filter:   make([]byte, 2048),
		hashFunc: []hashing.Hasher{keccak.NewKeccak(), blake2b.NewBlake2b(), fnv.NewFnv()},
	}
}

// Add sets the bits that correspond to the hashes of the data
func (b *Bloom) Add(data []byte) {
	res := getBitsIndexes(b, data)

	for i := range res {
		b.mutex.Lock()
		pos, bitMask := getBytePositionAndBitMask(res[i])

		b.filter[pos] |= bitMask
		b.mutex.Unlock()
	}

}

// MayContain checks if the bits that correspond to the hashes of the data are set.
// If all the bits are set, it returns true, otherwise it returns false
func (b *Bloom) MayContain(data []byte) bool {
	res := getBitsIndexes(b, data)

	for i := range res {
		pos, bitMask := getBytePositionAndBitMask(res[i])

		if b.filter[pos]&bitMask == 0 {
			return false
		}
	}
	return true
}

// Clear resets the bits of the bloom filter
func (b *Bloom) Clear() {
	for i := 0; i < len(b.filter); i++ {
		b.filter[i] = 0
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (b *Bloom) IsInterfaceNil() bool {
	return b == nil
}

// getBytePositionAndBitMask takes the index of a bit and returns the position of the byte in the filter
// that contains that index and the value of that byte with the bit set to 1.
func getBytePositionAndBitMask(index uint64) (pos uint64, val byte) {
	pos = index / 8
	bitNo := index % 8
	bitMask := byte(1 << bitNo)

	return pos, bitMask
}

// getBitsIndexes returns a slice which contains indexes that represent bits from the filter that have to be set.
func getBitsIndexes(b *Bloom, data []byte) []uint64 {
	var ch = make(chan uint64, len(b.hashFunc))
	var wg sync.WaitGroup
	var res []uint64

	wg.Add(len(b.hashFunc))
	for i := range b.hashFunc {
		go getBitIndexFromHash(b.hashFunc[i], data, len(b.filter), &wg, ch)
	}
	wg.Wait()

	for i := 0; i < len(b.hashFunc); i++ {
		res = append(res, <-ch)
	}

	return res

}

// getBitIndexFromHash hashes with the given hasher implementation the data received,
// then writes on the channel the index of the bit from the filter that has to be set.
func getBitIndexFromHash(h hashing.Hasher, data []byte, size int, wg *sync.WaitGroup, ch chan uint64) {

	hash := h.Compute(string(data))
	hash64 := binary.BigEndian.Uint64(hash)
	val := hash64 % uint64(size*bitsInByte)

	ch <- val
	wg.Done()

}

package bloom

import (
	"encoding/binary"
	"errors"
	"hash"
	"hash/fnv"
	"sync"

	"github.com/dchest/blake2b"
	"golang.org/x/crypto/sha3"
)

const (
	BitsInByte = 8
)

type Bloom struct {
	filter   []byte
	hashFunc []func() hash.Hash
}

type Result struct {
	index uint
	value byte
}

func NewFilter(size uint, h []func() hash.Hash) (*Bloom, error) {

	if size <= uint(len(h)) {
		return nil, errors.New("filter size is too low")
	}

	return &Bloom{
		filter:   make([]byte, size),
		hashFunc: h,
	}, nil
}

func NewDefaultFilter() (*Bloom, error) {
	return &Bloom{
		filter:   make([]byte, 2048),
		hashFunc: []func() hash.Hash{sha3.NewLegacyKeccak256, blake2b.New256, fnv.New128a},
	}, nil
}

func (b *Bloom) Add(data []byte) {
	var ch = make(chan Result, len(b.hashFunc))
	var wg sync.WaitGroup
	var res Result

	for i := 0; i < len(b.hashFunc); i++ {
		wg.Add(1)
		go getIndexAndValue(b.hashFunc[i](), data, b, &wg, ch)
	}
	wg.Wait()

	for i := 0; i < len(b.hashFunc); i++ {
		res = <-ch
		b.filter[res.index] = res.value
	}

}

func (b *Bloom) Test(data []byte) bool {
	var ch = make(chan Result, len(b.hashFunc))
	var wg sync.WaitGroup
	var res Result

	for i := 0; i < len(b.hashFunc); i++ {
		wg.Add(1)
		go getIndexAndValue(b.hashFunc[i](), data, b, &wg, ch)
	}
	wg.Wait()

	for i := 0; i < len(b.hashFunc); i++ {
		res = <-ch
		if b.filter[res.index] != res.value {
			return false
		}
	}
	return true
}

func (b *Bloom) Clear() {
	for i := 0; i < len(b.filter); i++ {
		b.filter[i] = 0
	}
}

func getIndexAndValue(h hash.Hash, data []byte, b *Bloom, wg *sync.WaitGroup, ch chan Result) {
	var res Result

	h.Write(data)
	hash64 := binary.BigEndian.Uint64(h.Sum(nil))
	val := hash64 % uint64(len(b.filter)*BitsInByte)

	byteNo := val / 8
	bitNo := val % 8
	byteValue := byte(1 << (7 - bitNo))
	value := byte(b.filter[byteNo] | byteValue)

	res.index = uint(byteNo)
	res.value = value

	ch <- res

	wg.Done()
}

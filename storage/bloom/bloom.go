package bloom

import (
	"encoding/binary"
	"errors"

	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/fnv"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
)

const (
	bitsInByte = 8
)

type Bloom struct {
	filter   []byte
	hashFunc []hashing.Hasher
}

type Result struct {
	index uint
	value byte
}

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

func NewDefaultFilter() (*Bloom, error) {
	return &Bloom{
		filter:   make([]byte, 2048),
		hashFunc: []hashing.Hasher{keccak.Keccak{}, blake2b.Blake2b{}, fnv.Fnv{}},
	}, nil
}

func (b *Bloom) Add(data []byte) {
	var ch = make(chan Result, len(b.hashFunc))
	var wg sync.WaitGroup
	var res Result

	wg.Add(len(b.hashFunc))
	for i := 0; i < len(b.hashFunc); i++ {
		go getIndexAndValue(b.hashFunc[i], data, b, &wg, ch)
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

	wg.Add(len(b.hashFunc))
	for i := 0; i < len(b.hashFunc); i++ {
		go getIndexAndValue(b.hashFunc[i], data, b, &wg, ch)
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

func getIndexAndValue(h hashing.Hasher, data []byte, b *Bloom, wg *sync.WaitGroup, ch chan Result) {
	var res Result

	hhh := h.Compute(string(data))
	hash64 := binary.BigEndian.Uint64(hhh)
	val := hash64 % uint64(len(b.filter)*bitsInByte)

	byteNo := val / 8
	bitNo := val % 8
	byteValue := byte(1 << (7 - bitNo))
	value := byte(b.filter[byteNo] | byteValue)

	res.index = uint(byteNo)
	res.value = value

	ch <- res

	wg.Done()
}

package bloom

import (
	"encoding/binary"
	"fmt"
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
	filter []byte
	size   uint64
	hash   []func() hash.Hash
}

type Result struct {
	index uint
	value byte
}

func NewFilter(size uint64, h []func() hash.Hash) Bloom {
	b := Bloom{
		filter: make([]byte, size),
		size:   size,
		hash:   h,
	}
	return b
}

func NewDefaultFilter() Bloom {

	h := []func() hash.Hash{sha3.NewLegacyKeccak256, blake2b.New256, fnv.New128a}

	b := Bloom{
		filter: make([]byte, 2048),
		size:   2048,
		hash:   h,
	}
	return b
}

func (b *Bloom) Add(data []byte) {
	var ch = make(chan Result, len(b.hash))
	var wg sync.WaitGroup
	var res Result

	for i := range b.hash {

		wg.Add(1)
		go getIndexAndValue(b.hash[i](), data, b, &wg, ch)
	}
	wg.Wait()

	for i := 0; i < len(b.hash); i++ {
		res = <-ch
		b.filter[res.index] = res.value
		fmt.Println(res.index)
	}

}

func (b *Bloom) Test(data []byte) bool {
	var ch = make(chan Result, len(b.hash))
	var wg sync.WaitGroup
	var res Result

	for i := range b.hash {
		wg.Add(1)
		go getIndexAndValue(b.hash[i](), data, b, &wg, ch)
	}
	wg.Wait()

	for i := 0; i < len(b.hash); i++ {
		res = <-ch
		if b.filter[res.index] != res.value {
			return false
		}
	}

	return true
}

func getIndexAndValue(h hash.Hash, data []byte, b *Bloom, wg *sync.WaitGroup, ch chan Result) {
	var res Result

	h.Write(data)
	hash := binary.BigEndian.Uint64(h.Sum(nil))
	val := uint(hash % (b.size * BitsInByte))

	byteNo := val / 8
	bitNo := val % 8
	byteValue := byte(1 << (7 - bitNo))
	value := byte(b.filter[byteNo] | byteValue)

	res.index = byteNo
	res.value = value

	ch <- res

	wg.Done()
}

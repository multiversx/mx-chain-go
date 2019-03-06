package state

import (
	"math/rand"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
)

type testInitializer struct {
	r *rand.Rand
}

func (ti *testInitializer) createDummyAddress() state.AddressContainer {
	buff := make([]byte, sha256.Sha256{}.Size())

	if ti.r == nil {
		ti.r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	ti.r.Read(buff)

	return state.NewAddress(buff)
}

func (ti *testInitializer) createMemUnit() storage.Storer {
	cache, _ := storage.NewCache(storage.LRUCache, 10)
	persist, _ := memorydb.New()

	unit, _ := storage.NewStorageUnit(cache, persist)
	return unit
}

func (ti *testInitializer) createDummyHexAddress(chars int) string {
	if chars < 1 {
		return ""
	}

	var characters = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

	rdm := rand.New(rand.NewSource(time.Now().UnixNano()))

	buff := make([]byte, chars)
	for i := 0; i < chars; i++ {
		buff[i] = characters[rdm.Int()%16]
	}

	return string(buff)
}

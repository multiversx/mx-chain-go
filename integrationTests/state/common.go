package state

import (
	"math/rand"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
)

func createDummyAddress() state.AddressContainer {
	buff := make([]byte, sha256.Sha256{}.Size())

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Read(buff)

	return state.NewAddress(buff)
}

func createMemUnit() storage.Storer {
	cache, _ := storage.NewCache(storage.LRUCache, 10)
	persist, _ := memorydb.New()

	unit, _ := storage.NewStorageUnit(cache, persist)
	return unit
}

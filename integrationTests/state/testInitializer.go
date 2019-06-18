package state

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/storageUnit"
)

func createDummyAddress() state.AddressContainer {
	buff := make([]byte, sha256.Sha256{}.Size())
	_, _ = rand.Reader.Read(buff)

	return state.NewAddress(buff)
}

func createMemUnit() storage.Storer {
	cache, _ := storageUnit.NewCache(storageUnit.LRUCache, 10, 1)
	persist, _ := memorydb.New()

	unit, _ := storageUnit.NewStorageUnit(cache, persist)
	return unit
}

func createDummyHexAddress(chars int) string {
	if chars < 1 {
		return ""
	}

	buff := make([]byte, chars/2)
	_, _ = rand.Reader.Read(buff)

	return hex.EncodeToString(buff)
}

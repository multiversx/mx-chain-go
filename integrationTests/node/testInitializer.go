package node

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	trie "github.com/ElrondNetwork/elrond-go-sandbox/integrationTests/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/storageUnit"
)

var testMarshalizer = &marshal.JsonMarshalizer{}

type accountFactory struct {
}

func (af *accountFactory) CreateAccount(address state.AddressContainer, tracker state.AccountTracker) (state.AccountHandler, error) {
	return state.NewAccount(address, tracker)
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

func createInMemoryShardAccountsDB() *state.AccountsDB {
	hasher := sha256.Sha256{}
	store := createMemUnit()

	pmt, _ := trie2.NewTrie(store, testMarshalizer, hasher)
	tr := trie.AdapterTrie{pmt}
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, testMarshalizer, &accountFactory{})

	return adb
}

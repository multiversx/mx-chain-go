package process

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

func createInMemoryAccountAdapter(
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	accountFactory state.AccountFactory,
) (state.AccountsAdapter, error) {

	trieStorage, err := trie.NewTrieStorageManagerWithoutPruning(createMemUnit())
	if err != nil {
		return nil, err
	}

	tr, err := trie.NewTrie(trieStorage, marshalizer, hasher)
	if err != nil {
		return nil, err
	}

	adb, err := state.NewAccountsDB(tr, hasher, marshalizer, accountFactory)
	if err != nil {
		return nil, err
	}

	return adb, nil
}

func createMemUnit() storage.Storer {
	cache, err := storageUnit.NewCache(storageUnit.LRUCache, 10, 1)
	if err != nil {
		log.Error("error creating cache for mem unit " + err.Error())
		return nil
	}

	unit, err := storageUnit.NewStorageUnit(cache, memorydb.New())
	if err != nil {
		log.Error("error creating unit " + err.Error())
		return nil
	}

	return unit
}

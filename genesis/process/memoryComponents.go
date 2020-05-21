package process

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

const maxTrieLevelInMemory = uint(5)

func createAccountAdapter(
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	accountFactory state.AccountFactory,
	trieStorage data.StorageManager,
) (state.AccountsAdapter, error) {
	tr, err := trie.NewTrie(trieStorage, marshalizer, hasher, maxTrieLevelInMemory)
	if err != nil {
		return nil, err
	}

	adb, err := state.NewAccountsDB(tr, hasher, marshalizer, accountFactory)
	if err != nil {
		return nil, err
	}

	return adb, nil
}

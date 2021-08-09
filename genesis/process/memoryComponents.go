package process

import (
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/disabled"
	"github.com/ElrondNetwork/elrond-go/trie"
)

const maxTrieLevelInMemory = uint(5)

func createAccountAdapter(
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	accountFactory state.AccountFactory,
	trieStorage common.StorageManager,
) (state.AccountsAdapter, error) {
	tr, err := trie.NewTrie(trieStorage, marshalizer, hasher, maxTrieLevelInMemory)
	if err != nil {
		return nil, err
	}

	adb, err := state.NewAccountsDB(
		tr,
		hasher,
		marshalizer,
		accountFactory,
		disabled.NewDisabledStoragePruningManager(),
	)
	if err != nil {
		return nil, err
	}

	return adb, nil
}

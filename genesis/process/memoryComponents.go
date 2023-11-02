package process

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	commonDisabled "github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/disabled"
	"github.com/multiversx/mx-chain-go/trie"
)

const maxTrieLevelInMemory = uint(5)

func createAccountAdapter(
	marshaller marshal.Marshalizer,
	hasher hashing.Hasher,
	accountFactory state.AccountFactory,
	trieStorage common.StorageManager,
	addressConverter core.PubkeyConverter,
	enableEpochsHandler common.EnableEpochsHandler,
) (state.AccountsAdapter, error) {
	tr, err := trie.NewTrie(trieStorage, marshaller, hasher, enableEpochsHandler, maxTrieLevelInMemory)
	if err != nil {
		return nil, err
	}

	args := state.ArgsAccountsDB{
		Trie:                  tr,
		Hasher:                hasher,
		Marshaller:            marshaller,
		AccountFactory:        accountFactory,
		StoragePruningManager: disabled.NewDisabledStoragePruningManager(),
		ProcessingMode:        common.Normal,
		ProcessStatusHandler:  commonDisabled.NewProcessStatusHandler(),
		AppStatusHandler:      commonDisabled.NewAppStatusHandler(),
		AddressConverter:      addressConverter,
		EnableEpochsHandler:   enableEpochsHandler,
	}

	adb, err := state.NewAccountsDB(args)
	if err != nil {
		return nil, err
	}

	return adb, nil
}

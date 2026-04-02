package process

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	disabledState "github.com/multiversx/mx-chain-go/state/disabled"
	factoryState "github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/disabled"
	"github.com/multiversx/mx-chain-go/state/triesHolder"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/collapseManager"
)

func createAccountAdapter(
	coreComp coreComponentsHandler,
	trieStorage common.StorageManager,
	addressConverter core.PubkeyConverter,
) (state.AccountsAdapter, error) {
	tr, err := trie.NewTrie(trieStorage, coreComp.InternalMarshalizer(), coreComp.Hasher(), coreComp.EnableEpochsHandler(), collapseManager.NewDisabledCollapseManager())
	if err != nil {
		return nil, err
	}

	tenMbSize := uint64(10485760)
	dth, err := triesHolder.NewDataTriesHolder(tenMbSize)
	if err != nil {
		return nil, err
	}

	argsAccCreator := factoryState.ArgsAccountCreator{
		Hasher:                 coreComp.Hasher(),
		Marshaller:             coreComp.InternalMarshalizer(),
		EnableEpochsHandler:    coreComp.EnableEpochsHandler(),
		StateAccessesCollector: disabledState.NewDisabledStateAccessesCollector(),
		DataTriesHolder:        dth,
		DataTrieCreator:        tr,
	}
	accCreator, err := factoryState.NewAccountCreator(argsAccCreator)
	if err != nil {
		return nil, err
	}

	args := state.ArgsAccountsDB{
		Trie:                   tr,
		Hasher:                 coreComp.Hasher(),
		Marshaller:             coreComp.InternalMarshalizer(),
		AccountFactory:         accCreator,
		StoragePruningManager:  disabled.NewDisabledStoragePruningManager(),
		AddressConverter:       addressConverter,
		SnapshotsManager:       disabledState.NewDisabledSnapshotsManager(),
		StateAccessesCollector: disabledState.NewDisabledStateAccessesCollector(),
		DataTriesHolder:        dth,
	}

	adb, err := state.NewAccountsDB(args)
	if err != nil {
		return nil, err
	}

	return adb, nil
}

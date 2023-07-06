package state

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state/iteratorChannelsProvider"
	"github.com/multiversx/mx-chain-go/state/stateMetrics"
)

// PeerAccountsDB will save and synchronize data from peer processor, plus will synchronize with nodesCoordinator
type PeerAccountsDB struct {
	*AccountsDB
}

// NewPeerAccountsDB creates a new account manager
func NewPeerAccountsDB(args ArgsAccountsDB) (*PeerAccountsDB, error) {
	err := checkArgsAccountsDB(args)
	if err != nil {
		return nil, err
	}

	argStateMetrics := stateMetrics.ArgsStateMetrics{
		SnapshotInProgressKey:   common.MetricPeersSnapshotInProgress,
		LastSnapshotDurationKey: common.MetricLastPeersSnapshotDurationSec,
		SnapshotMessage:         stateMetrics.PeerTrieSnapshotMsg,
	}
	sm, err := stateMetrics.NewStateMetrics(argStateMetrics, args.AppStatusHandler)
	if err != nil {
		return nil, err
	}

	argsSnapshotsManager := ArgsNewSnapshotsManager{
		ShouldSerializeSnapshots: args.ShouldSerializeSnapshots,
		ProcessingMode:           args.ProcessingMode,
		Marshaller:               args.Marshaller,
		AddressConverter:         args.AddressConverter,
		ProcessStatusHandler:     args.ProcessStatusHandler,
		StateMetrics:             sm,
		ChannelsProvider:         iteratorChannelsProvider.NewPeerStateIteratorChannelsProvider(),
	}
	snapshotManager, err := NewSnapshotsManager(argsSnapshotsManager)
	if err != nil {
		return nil, err
	}

	adb := &PeerAccountsDB{
		AccountsDB: createAccountsDb(args, snapshotManager),
	}

	return adb, nil
}

// MarkSnapshotDone will mark that the snapshot process has been completed
func (adb *PeerAccountsDB) MarkSnapshotDone() {
	trieStorageManager, epoch, err := adb.getTrieStorageManagerAndLatestEpoch(adb.getMainTrie())
	if err != nil {
		log.Error("MarkSnapshotDone error", "err", err.Error())
		return
	}

	err = trieStorageManager.PutInEpochWithoutCache([]byte(common.ActiveDBKey), []byte(common.ActiveDBVal), epoch)
	handleLoggingWhenError("error while putting active DB value into main storer", err)
}

// RecreateAllTries recreates all the tries from the accounts DB
func (adb *PeerAccountsDB) RecreateAllTries(rootHash []byte) (map[string]common.Trie, error) {
	return adb.recreateMainTrie(rootHash)
}

// IsInterfaceNil returns true if there is no value under the interface
func (adb *PeerAccountsDB) IsInterfaceNil() bool {
	return adb == nil
}

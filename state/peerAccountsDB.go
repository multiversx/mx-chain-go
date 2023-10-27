package state

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state/iteratorChannelsProvider"
	"github.com/multiversx/mx-chain-go/state/lastSnapshotMarker"
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
		AccountFactory:           args.AccountFactory,
		LastSnapshotMarker:       lastSnapshotMarker.NewLastSnapshotMarker(),
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

// GetPeerAccountAndReturnIfNew returns the peer account and a flag indicating if the account is new
func GetPeerAccountAndReturnIfNew(adb AccountsAdapter, address []byte) (PeerAccountHandler, bool, error) {
	var err error

	newAccount := false
	account, _ := adb.GetExistingAccount(address)
	if check.IfNil(account) {
		newAccount = true
		account, err = adb.LoadAccount(address)
		if err != nil {
			return nil, false, err
		}
	}

	peerAcc, ok := account.(PeerAccountHandler)
	if !ok {
		return nil, false, ErrWrongTypeAssertion
	}

	return peerAcc, newAccount, nil
}

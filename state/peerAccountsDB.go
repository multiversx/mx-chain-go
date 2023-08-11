package state

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
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

	adb := &PeerAccountsDB{
		AccountsDB: createAccountsDb(args),
	}

	args.AppStatusHandler.SetUInt64Value(common.MetricPeersSnapshotInProgress, 0)

	return adb, nil
}

// StartSnapshotIfNeeded starts the snapshot if the previous snapshot process was not fully completed
func (adb *PeerAccountsDB) StartSnapshotIfNeeded() error {
	return startSnapshotIfNeeded(adb, adb.getTrieSyncer(), adb.getMainTrie().GetStorageManager(), adb.processingMode)
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

// SnapshotState triggers the snapshotting process of the state trie
func (adb *PeerAccountsDB) SnapshotState(rootHash []byte) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	trieStorageManager, epoch, shouldTakeSnapshot := adb.prepareSnapshot(rootHash)
	if !shouldTakeSnapshot {
		return
	}

	log.Info("starting snapshot peer trie", "rootHash", rootHash, "epoch", epoch)
	missingNodesChannel := make(chan []byte, missingNodesChannelSize)
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: nil,
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	stats := newSnapshotStatistics(0, 1)
	stats.NewSnapshotStarted()

	peerAccountsMetrics := &accountMetrics{
		snapshotInProgressKey:   common.MetricPeersSnapshotInProgress,
		lastSnapshotDurationKey: common.MetricLastPeersSnapshotDurationSec,
		snapshotMessage:         peerTrieSnapshotMsg,
	}
	adb.updateMetricsOnSnapshotStart(peerAccountsMetrics)

	trieStorageManager.TakeSnapshot("", rootHash, rootHash, iteratorChannels, missingNodesChannel, stats, epoch)

	go adb.syncMissingNodes(missingNodesChannel, iteratorChannels.ErrChan, stats, adb.trieSyncer)

	go adb.processSnapshotCompletion(stats, trieStorageManager, missingNodesChannel, iteratorChannels.ErrChan, rootHash, peerAccountsMetrics, epoch)

	adb.waitForCompletionIfAppropriate(stats)
}

// SetStateCheckpoint triggers the checkpointing process of the state trie
func (adb *PeerAccountsDB) SetStateCheckpoint(rootHash []byte) {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()

	log.Trace("peerAccountsDB.SetStateCheckpoint", "root hash", rootHash)
	trieStorageManager := adb.mainTrie.GetStorageManager()

	missingNodesChannel := make(chan []byte, missingNodesChannelSize)
	stats := newSnapshotStatistics(0, 1)

	trieStorageManager.EnterPruningBufferingMode()
	stats.NewSnapshotStarted()
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: nil,
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	trieStorageManager.SetCheckpoint(rootHash, rootHash, iteratorChannels, missingNodesChannel, stats)

	go adb.syncMissingNodes(missingNodesChannel, iteratorChannels.ErrChan, stats, adb.trieSyncer)

	// TODO decide if we need to take some actions whenever we hit an error that occurred in the checkpoint process
	//  that will be present in the errChan var
	go adb.finishSnapshotOperation(rootHash, stats, missingNodesChannel, "setStateCheckpoint peer trie", trieStorageManager)

	adb.waitForCompletionIfAppropriate(stats)
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

package state

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/common"
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
		AccountsDB: getAccountsDb(args),
	}

	trieStorageManager := adb.mainTrie.GetStorageManager()
	val, err := trieStorageManager.GetFromCurrentEpoch([]byte(common.ActiveDBKey))
	if err != nil || !bytes.Equal(val, []byte(common.ActiveDBVal)) {
		startSnapshotAfterRestart(adb, args)
	}

	return adb, nil
}

// MarkSnapshotDone will mark that the snapshot process has been completed
func (adb *PeerAccountsDB) MarkSnapshotDone() {
	trieStorageManager, epoch, err := adb.getTrieStorageManagerAndLatestEpoch()
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
	errChan := make(chan error, 1)
	stats := newSnapshotStatistics(0)
	stats.NewSnapshotStarted()
	trieStorageManager.TakeSnapshot(rootHash, rootHash, nil, errChan, stats, epoch)
	trieStorageManager.ExitPruningBufferingMode()

	go adb.markActiveDBAfterSnapshot(stats, errChan, rootHash, "snapshotState peer trie", epoch)

	adb.waitForCompletionIfRunningInImportDB(stats)
}

// SetStateCheckpoint triggers the checkpointing process of the state trie
func (adb *PeerAccountsDB) SetStateCheckpoint(rootHash []byte) {
	log.Trace("peerAccountsDB.SetStateCheckpoint", "root hash", rootHash)
	trieStorageManager := adb.mainTrie.GetStorageManager()

	stats := newSnapshotStatistics(0)

	trieStorageManager.EnterPruningBufferingMode()
	stats.NewSnapshotStarted()
	errChan := make(chan error, 1)
	trieStorageManager.SetCheckpoint(rootHash, rootHash, nil, errChan, stats)
	trieStorageManager.ExitPruningBufferingMode()

	// TODO decide if we need to take some actions whenever we hit an error that occurred in the checkpoint process
	//  that will be present in the errChan var
	go stats.PrintStats("setStateCheckpoint peer trie", rootHash)

	adb.waitForCompletionIfRunningInImportDB(stats)
}

// RecreateAllTries recreates all the tries from the accounts DB
func (adb *PeerAccountsDB) RecreateAllTries(rootHash []byte) (map[string]common.Trie, error) {
	return adb.recreateMainTrie(rootHash)
}

// IsInterfaceNil returns true if there is no value under the interface
func (adb *PeerAccountsDB) IsInterfaceNil() bool {
	return adb == nil
}
